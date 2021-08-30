package buffer

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"vitess.io/vitess/go/vt/log"
	"vitess.io/vitess/go/vt/logutil"
	topodatapb "vitess.io/vitess/go/vt/proto/topodata"
	"vitess.io/vitess/go/vt/topo/topoproto"
	"vitess.io/vitess/go/vt/vterrors"
)

// shardBuffer buffers requests during a failover for a particular shard.
// The object will be reused across failovers. If no failover is currently in
// progress, the state is "IDLE".
//
// Note that this object is accessed concurrently by multiple threads:
// - vtgate request threads
// - discovery.HealthCheck listener execution thread
// - timeout thread (timeout_thread.go) to evict too old buffered requests
// - drain() thread
type shardBufferHC struct {
	// Immutable fields set at construction.
	buf      *HealthCheckBuffer
	mode     bufferMode
	keyspace string
	shard    string

	// statsKey is used to update the stats variables.
	statsKey []string
	// statsKeyJoined is all elements of "statsKey" in one string, joined by ".".
	statsKeyJoined string
	logTooRecent   *logutil.ThrottledLogger

	// mu guards the fields below.
	mu    sync.RWMutex
	state bufferState
	// queue is the list of buffered requests (ordered by arrival).
	queue []*entry
	// externallyReparented is the maximum value of all seen
	// "StreamHealthResponse.TabletexternallyReparentedTimestamp" values across
	// all PRIMARY tablets of this shard.
	// In practice, it is a) the last time the shard was reparented or b) the last
	// time the TabletExternallyReparented RPC was called on the tablet to confirm
	// that the tablet is the current PRIMARY.
	// We assume the value is a Unix timestamp in seconds.
	externallyReparented int64
	// lastStart is the last time we saw the start of a failover.
	lastStart time.Time
	// lastEnd is the last time we saw the end of a failover.
	lastEnd time.Time
	// lastReparent is the last time we saw that the tablet alias of the PRIMARY
	// changed i.e. we definitely reparented to a different tablet.
	lastReparent time.Time
	// currentPrimary is tracked to determine when to update "lastReparent".
	currentPrimary *topodatapb.TabletAlias
	// timeoutThread will be set while a failover is in progress and the object is
	// in the BUFFERING state.
	timeoutThread *timeoutThread
	// wg tracks all pending Go routines. waitForShutdown() will use this field to
	// block on them.
	wg sync.WaitGroup
}

func newShardBufferHealthCheck(buf *HealthCheckBuffer, mode bufferMode, keyspace, shard string) *shardBufferHC {
	statsKey := []string{keyspace, shard}
	initVariablesForShard(statsKey)

	return &shardBufferHC{
		buf:            buf,
		mode:           mode,
		keyspace:       keyspace,
		shard:          shard,
		statsKey:       statsKey,
		statsKeyJoined: fmt.Sprintf("%s.%s", keyspace, shard),
		logTooRecent:   logutil.NewThrottledLogger(fmt.Sprintf("FailoverTooRecent-%v", topoproto.KeyspaceShardString(keyspace, shard)), 5*time.Second),
		state:          stateIdle,
	}
}

func (sb *shardBufferHC) timeNow() time.Time {
	return sb.buf.config.now()
}

// disabled returns true if neither buffering nor the dry-run mode is enabled.
func (sb *shardBufferHC) disabled() bool {
	return sb.mode == bufferDisabled
}

func (sb *shardBufferHC) waitForFailoverEnd(ctx context.Context, keyspace, shard string, err error) (RetryDoneFunc, error) {
	// We assume if err != nil then it's always caused by a failover.
	// Other errors must be filtered at higher layers.
	failoverDetected := err != nil

	// Fast path (read lock): Check if we should NOT buffer a request.
	sb.mu.RLock()
	if !sb.shouldBufferLocked(failoverDetected) {
		// No buffering required. Return early.
		sb.mu.RUnlock()
		return nil, nil
	}
	sb.mu.RUnlock()

	// Buffering required. Acquire write lock.
	sb.mu.Lock()
	// Re-check state because it could have changed in the meantime.
	if !sb.shouldBufferLocked(failoverDetected) {
		// Buffering no longer required. Return early.
		sb.mu.Unlock()
		return nil, nil
	}

	// Start buffering if failover is not detected yet.
	if sb.state == stateIdle {
		// Do not buffer if last failover is too recent. This is the case if:
		// a) buffering was stopped recently
		// OR
		// b) we did not buffer, but observed a reparent very recently
		now := sb.timeNow()
		minTimeBetweenFailovers := sb.buf.config.MinTimeBetweenFailovers

		// a) Buffering was stopped recently.
		// This can happen when we stop buffering while MySQL is not ready yet
		// (read-only mode is not cleared yet on the new primary).
		lastBufferingStopped := now.Sub(sb.lastEnd)
		if !sb.lastEnd.IsZero() && lastBufferingStopped < minTimeBetweenFailovers {
			sb.mu.Unlock()
			msg := "NOT starting buffering"
			if sb.mode == bufferDryRun {
				msg = "Dry-run: Would NOT have started buffering"
			}

			sb.logTooRecent.Infof("%v for shard: %s because the last failover which triggered buffering is too recent (%v < %v)."+
				" (A failover was detected by this seen error: %v.)",
				msg, topoproto.KeyspaceShardString(keyspace, shard), lastBufferingStopped, minTimeBetweenFailovers, err)

			statsKeyWithReason := append(sb.statsKey, string(skippedLastFailoverTooRecent))
			requestsSkipped.Add(statsKeyWithReason, 1)
			return nil, nil
		}

		// b) The PRIMARY was reparented recently (but we did not buffer it.)
		// This can happen when we see the end of the reparent *before* the first
		// request failure caused by the reparent. This is possible if the QPS is
		// very low. If we do not skip buffering here, we would start buffering but
		// not stop because we already observed the promotion of the new primary.
		lastReparentAgo := now.Sub(sb.lastReparent)
		if !sb.lastReparent.IsZero() && lastReparentAgo < minTimeBetweenFailovers {
			sb.mu.Unlock()
			msg := "NOT starting buffering"
			if sb.mode == bufferDryRun {
				msg = "Dry-run: Would NOT have started buffering"
			}

			sb.logTooRecent.Infof("%v for shard: %s because the last reparent is too recent (%v < %v)."+
				" (A failover was detected by this seen error: %v.)",
				msg, topoproto.KeyspaceShardString(keyspace, shard), lastReparentAgo, minTimeBetweenFailovers, err)

			statsKeyWithReason := append(sb.statsKey, string(skippedLastReparentTooRecent))
			requestsSkipped.Add(statsKeyWithReason, 1)
			return nil, nil
		}

		sb.startBufferingLocked(err)
	}

	if sb.mode == bufferDryRun {
		sb.mu.Unlock()
		// Dry-run. Do not actually buffer the request and return early.
		lastRequestsDryRunMax.Add(sb.statsKey, 1)
		requestsBufferedDryRun.Add(sb.statsKey, 1)
		return nil, nil
	}

	// Buffer request.
	entry, err := sb.bufferRequestLocked(ctx)
	sb.mu.Unlock()
	if err != nil {
		return nil, err
	}
	return sb.wait(ctx, entry)
}

// shouldBufferLocked returns true if the current request should be buffered
// (based on the current state and whether the request detected a failover).
func (sb *shardBufferHC) shouldBufferLocked(failoverDetected bool) bool {
	switch s := sb.state; {
	case s == stateIdle && !failoverDetected:
		// No failover in progress.
		return false
	case s == stateIdle && failoverDetected:
		// Not buffering yet, but new failover detected.
		return true
	case s == stateBuffering:
		// Failover in progress.
		return true
	case s == stateDraining && !failoverDetected:
		// Draining. Non-failover related requests can pass through.
		return false
	case s == stateDraining && failoverDetected:
		// Possible race between request which saw failover-related error and the
		// end of the failover. Do not buffer and let vtgate retry immediately.
		return false
	}
	panic("BUG: All possible states must be covered by the switch expression above.")
}

func (sb *shardBufferHC) startBufferingLocked(err error) {
	// Reset monitoring data from previous failover.
	lastRequestsInFlightMax.Set(sb.statsKey, 0)
	lastRequestsDryRunMax.Set(sb.statsKey, 0)
	failoverDurationSumMs.Reset(sb.statsKey)

	sb.lastStart = sb.timeNow()
	sb.logErrorIfStateNotLocked(stateIdle)
	sb.state = stateBuffering
	sb.queue = make([]*entry, 0)

	sb.timeoutThread = newTimeoutThread(sb, sb.buf.config.MaxFailoverDuration)
	sb.timeoutThread.start()
	msg := "Starting buffering"
	if sb.mode == bufferDryRun {
		msg = "Dry-run: Would have started buffering"
	}
	starts.Add(sb.statsKey, 1)
	log.Infof("%v for shard: %s (window: %v, size: %v, max failover duration: %v) (A failover was detected by this seen error: %v.)",
		msg,
		topoproto.KeyspaceShardString(sb.keyspace, sb.shard),
		sb.buf.config.Window,
		sb.buf.config.Size,
		sb.buf.config.MaxFailoverDuration,
		err,
	)
}

// logErrorIfStateNotLocked logs an error if the current state is not "state".
// We do not panic/crash the process here because it is expected that a wrong
// state is less severe than (potentially) crash-looping all vtgates.
// Note: The prefix "Locked" is not related to the state. Instead, it stresses
// that "sb.mu" must be locked before calling the method.
func (sb *shardBufferHC) logErrorIfStateNotLocked(state bufferState) {
	if sb.state != state {
		log.Errorf("BUG: Buffer state should be '%v' and not '%v'. Full state of buffer object: %#v Stacktrace:\n%s", state, sb.state, sb, debug.Stack())
	}
}

// bufferRequest creates a new entry in the queue for a request which
// should be buffered.
// It returns *entry which can be used as input for shardBuffer.cancel(). This
// is useful for canceled RPCs (e.g. due to deadline exceeded) which want to
// give up their spot in the buffer. It also holds the "bufferCancel" function.
// If buffering fails e.g. due to a full buffer, an error is returned.
func (sb *shardBufferHC) bufferRequestLocked(ctx context.Context) (*entry, error) {
	if !sb.buf.bufferSizeSema.TryAcquire() {
		// Buffer is full. Evict the oldest entry and buffer this request instead.
		if len(sb.queue) == 0 {
			// Overall buffer is full, but this shard's queue is empty. That means
			// there is at least one other shard failing over as well which consumes
			// the whole buffer.
			statsKeyWithReason := append(sb.statsKey, string(skippedBufferFull))
			requestsSkipped.Add(statsKeyWithReason, 1)
			return nil, bufferFullError
		}

		e := sb.queue[0]
		// Evict the entry. Do not release its slot in the buffer and reuse it for
		// this new request.
		// NOTE: We keep the lock to avoid racing with drain().
		// NOTE: We're not waiting until the request finishes and instead reuse its
		// slot immediately, i.e. the number of evicted requests + drained requests
		// can be bigger than the buffer size.
		sb.unblockAndWait(e, entryEvictedError, false /* releaseSlot */, false /* blockingWait */)
		sb.queue = sb.queue[1:]
		statsKeyWithReason := append(sb.statsKey, evictedBufferFull)
		requestsEvicted.Add(statsKeyWithReason, 1)
	}

	e := &entry{
		done:     make(chan struct{}),
		deadline: sb.timeNow().Add(sb.buf.config.Window),
	}
	e.bufferCtx, e.bufferCancel = context.WithCancel(ctx)
	sb.queue = append(sb.queue, e)

	if max := lastRequestsInFlightMax.Counts()[sb.statsKeyJoined]; max < int64(len(sb.queue)) {
		lastRequestsInFlightMax.Set(sb.statsKey, int64(len(sb.queue)))
	}
	requestsBuffered.Add(sb.statsKey, 1)

	if len(sb.queue) == 1 {
		sb.timeoutThread.notifyQueueNotEmpty()
	}
	return e, nil
}

// unblockAndWait unblocks a blocked request.
// If releaseSlot is true, the buffer semaphore will be decreased by 1 when
// the request retried and finished.
// If blockingWait is true, this call will block until the request retried and
// finished. This mode is used during the drain (to avoid flooding the primary)
// while the non-blocking mode is used when a) evicting a request (e.g. because
// the buffer is full or it exceeded the buffering window) or b) when the
// request was canceled from outside and we removed it.
func (sb *shardBufferHC) unblockAndWait(e *entry, err error, releaseSlot, blockingWait bool) {
	// Set error such that the request will see it.
	e.err = err
	// Tell blocked request to stop waiting.
	close(e.done)

	if blockingWait {
		sb.waitForRequestFinish(e, releaseSlot, false /* async */)
	} else {
		sb.wg.Add(1)
		go sb.waitForRequestFinish(e, releaseSlot, true /* async */)
	}
}

func (sb *shardBufferHC) waitForRequestFinish(e *entry, releaseSlot, async bool) {
	if async {
		defer sb.wg.Done()
	}

	// Wait for unblocked request to finish.
	<-e.bufferCtx.Done()

	// Release the slot to the buffer.
	// NOTE: We always wait for the request first, even if the calling code like
	// the buffer full eviction or the timeout thread does not block on us.
	// This way, the request's slot can only be reused after the request finished.
	if releaseSlot {
		sb.buf.bufferSizeSema.Release()
	}
}

// wait blocks while the request is buffered during the failover.
// See Buffer.WaitForFailoverEnd() for the API contract of the return values.
func (sb *shardBufferHC) wait(ctx context.Context, e *entry) (RetryDoneFunc, error) {
	select {
	case <-ctx.Done():
		sb.remove(e)
		return nil, vterrors.Errorf(vterrors.Code(contextCanceledError), "%v: %v", contextCanceledError, ctx.Err())
	case <-e.done:
		return e.bufferCancel, e.err
	}
}

// oldestEntry returns the head of the queue or nil if the queue is empty.
func (sb *shardBufferHC) oldestEntry() *entry {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if len(sb.queue) > 0 {
		return sb.queue[0]
	}
	return nil
}

// evictOldestEntry is used by timeoutThread to evict the head entry of the
// queue if it exceeded its buffering window.
func (sb *shardBufferHC) evictOldestEntry(e *entry) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if len(sb.queue) == 0 || e != sb.queue[0] {
		// Entry is already removed e.g. by remove(). Ignore it.
		return
	}

	// Evict the entry.
	//
	// NOTE: We're not waiting for the request to finish in order to unblock the
	// timeout thread as fast as possible. However, the slot of the evicted
	// request is only returned after it has finished i.e. the buffer may stay
	// full in the meantime. This is a design tradeoff to keep things simple and
	// avoid additional pressure on the primary tablet.
	sb.unblockAndWait(e, nil /* err */, true /* releaseSlot */, false /* blockingWait */)
	sb.queue = sb.queue[1:]
	statsKeyWithReason := append(sb.statsKey, evictedWindowExceeded)
	requestsEvicted.Add(statsKeyWithReason, 1)
}

// remove must be called when the request was canceled from outside and not
// internally.
func (sb *shardBufferHC) remove(toRemove *entry) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if sb.queue == nil {
		// Queue is cleared because we're already in the DRAIN phase.
		return
	}

	// If entry is still in the queue, delete it and cancel it internally.
	for i, e := range sb.queue {
		if e == toRemove {
			// Delete entry at index "i" from slice.
			sb.queue = append(sb.queue[:i], sb.queue[i+1:]...)

			// Cancel the entry's "bufferCtx".
			// The usual drain or eviction code would unblock the request and then
			// wait for the "bufferCtx" to be done.
			// But this code path is different because it's going to return an error
			// to the request and not the "e.bufferCancel" function i.e. the request
			// cannot cancel the "bufferCtx" itself.
			// Therefore, we call "e.bufferCancel". This also avoids that the
			// context's Go routine could leak.
			e.bufferCancel()
			// Release the buffer slot and close the "e.done" channel.
			// By closing "e.done", we finish it explicitly and timeoutThread will
			// find out about it as well.
			sb.unblockAndWait(e, nil /* err */, true /* releaseSlot */, false /* blockingWait */)

			// Track it as "ContextDone" eviction.
			statsKeyWithReason := append(sb.statsKey, string(evictedContextDone))
			requestsEvicted.Add(statsKeyWithReason, 1)
			return
		}
	}

	// Entry was already removed. Keep the queue as it is.
}

func (sb *shardBufferHC) recordExternallyReparentedTimestamp(timestamp int64, alias *topodatapb.TabletAlias) {
	// Fast path (read lock): Check if new timestamp is higher.
	sb.mu.RLock()
	if timestamp <= sb.externallyReparented {
		// Do nothing. Equal values are reported if the primary has not changed.
		// Smaller values can be reported during the failover by the old primary
		// after the new primary already took over.
		sb.mu.RUnlock()
		return
	}
	sb.mu.RUnlock()

	// New timestamp is higher. Stop buffering if running.
	sb.mu.Lock()
	defer sb.mu.Unlock()

	// Re-check value after acquiring write lock.
	if timestamp <= sb.externallyReparented {
		return
	}

	sb.externallyReparented = timestamp
	if !topoproto.TabletAliasEqual(alias, sb.currentPrimary) {
		if sb.currentPrimary != nil {
			sb.lastReparent = sb.timeNow()
		}
		sb.currentPrimary = alias
	}
	sb.stopBufferingLocked(stopFailoverEndDetected, "failover end detected")
}

func (sb *shardBufferHC) stopBufferingDueToMaxDuration() {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	sb.stopBufferingLocked(stopMaxFailoverDurationExceeded,
		fmt.Sprintf("stopping buffering because failover did not finish in time (%v)", sb.buf.config.MaxFailoverDuration))
}

func (sb *shardBufferHC) stopBufferingLocked(reason stopReason, details string) {
	if sb.state != stateBuffering {
		return
	}

	// Stop buffering.
	sb.lastEnd = sb.timeNow()
	d := sb.lastEnd.Sub(sb.lastStart)

	statsKeyWithReason := append(sb.statsKey, string(reason))
	stops.Add(statsKeyWithReason, 1)

	lastFailoverDurationMs.Set(sb.statsKey, int64(d/time.Millisecond))
	failoverDurationSumMs.Add(sb.statsKey, int64(d/time.Millisecond))
	if sb.mode == bufferDryRun {
		utilDryRunMax := int64(
			float64(lastRequestsDryRunMax.Counts()[sb.statsKeyJoined]) / float64(sb.buf.config.Size) * 100.0)
		utilizationDryRunSum.Add(sb.statsKey, utilDryRunMax)
	} else {
		utilMax := int64(
			float64(lastRequestsInFlightMax.Counts()[sb.statsKeyJoined]) / float64(sb.buf.config.Size) * 100.0)
		utilizationSum.Add(sb.statsKey, utilMax)
	}

	sb.logErrorIfStateNotLocked(stateBuffering)
	sb.state = stateDraining
	q := sb.queue
	// Clear the queue such that remove(), oldestEntry() and evictOldestEntry()
	// will not work on obsolete data.
	sb.queue = nil

	msg := "Stopping buffering"
	if sb.mode == bufferDryRun {
		msg = "Dry-run: Would have stopped buffering"
	}
	log.Infof("%v for shard: %s after: %.1f seconds due to: %v. Draining %d buffered requests now.", msg, topoproto.KeyspaceShardString(sb.keyspace, sb.shard), d.Seconds(), details, len(q))

	// Start the drain. (Use a new Go routine to release the lock.)
	sb.wg.Add(1)
	go sb.drain(q)
}

func (sb *shardBufferHC) drain(q []*entry) {
	defer sb.wg.Done()

	// stop must be called outside of the lock because the thread may access
	// shardBuffer as well e.g. to get the current oldest entry.
	sb.timeoutThread.stop()

	start := sb.timeNow()
	// TODO(mberlin): Parallelize the drain by pumping the data through a channel.
	for _, e := range q {
		sb.unblockAndWait(e, nil /* err */, true /* releaseSlot */, true /* blockingWait */)
	}
	d := sb.timeNow().Sub(start)
	log.Infof("Draining finished for shard: %s Took: %v for: %d requests.", topoproto.KeyspaceShardString(sb.keyspace, sb.shard), d, len(q))
	requestsDrained.Add(sb.statsKey, int64(len(q)))

	// Draining is done. Change state from "draining" to "idle".
	sb.mu.Lock()
	defer sb.mu.Unlock()
	sb.logErrorIfStateNotLocked(stateDraining)
	sb.state = stateIdle
	sb.timeoutThread = nil
}

func (sb *shardBufferHC) shutdown() {
	sb.mu.Lock()
	sb.stopBufferingLocked(stopShutdown, "shutdown")
	sb.mu.Unlock()
}

func (sb *shardBufferHC) waitForShutdown() {
	sb.wg.Wait()
}

// sizeForTesting is used by the unit test only to find out the current number
// of buffered requests.
// TODO(mberlin): Remove this if we add a more general statistics reporting.
func (sb *shardBufferHC) testGetSize() int {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	return len(sb.queue)
}

// stateForTesting is used by unit tests only to probe the current state.
func (sb *shardBufferHC) testGetState() bufferState {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	return sb.state
}

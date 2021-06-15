/*
Copyright 2020 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package vtorc

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	"vitess.io/vitess/go/mysql"
	"vitess.io/vitess/go/sqltypes"

	"vitess.io/vitess/go/vt/topo/topoproto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"vitess.io/vitess/go/json2"
	"vitess.io/vitess/go/test/endtoend/cluster"
	"vitess.io/vitess/go/vt/log"
	querypb "vitess.io/vitess/go/vt/proto/query"
	topodatapb "vitess.io/vitess/go/vt/proto/topodata"
)

var (
	clusterInstance *cluster.LocalProcessCluster
	replicaTablets  []*cluster.Vttablet
	rdonlyTablets   []*cluster.Vttablet
	uidBase         = 100
)

const (
	keyspaceName = "ks"
	shardName    = "0"
	hostname     = "localhost"
	cell1        = "zone1"
	numReplicas  = 3
	numRdonly    = 1
)

// createClusterAndStartTopo starts the cluster and topology service
func createClusterAndStartTopo() error {
	clusterInstance = cluster.NewCluster(cell1, hostname)

	// Start topo server
	err := clusterInstance.StartTopo()
	if err != nil {
		return err
	}

	// create the vttablets
	err = createVttablets()
	if err != nil {
		return err
	}

	return err
}

// createVttablets is used to create the vttablets for all the tests
func createVttablets() error {
	keyspace := &cluster.Keyspace{Name: keyspaceName}
	shard0 := &cluster.Shard{Name: shardName}

	// creating tablets by hand instead of using StartKeyspace because we don't want to call InitShardMaster
	var tablets []*cluster.Vttablet
	for i := 0; i < numReplicas; i++ {
		vttabletInstance := clusterInstance.NewVttabletInstance("replica", uidBase, cell1)
		uidBase++
		tablets = append(tablets, vttabletInstance)
		replicaTablets = append(replicaTablets, vttabletInstance)
	}
	for i := 0; i < numRdonly; i++ {
		vttabletInstance := clusterInstance.NewVttabletInstance("rdonly", uidBase, cell1)
		uidBase++
		tablets = append(tablets, vttabletInstance)
		rdonlyTablets = append(rdonlyTablets, vttabletInstance)
	}
	clusterInstance.VtTabletExtraArgs = []string{
		"-lock_tables_timeout", "5s",
		"-disable_active_reparents",
	}
	// Initialize Cluster
	shard0.Vttablets = tablets
	err := clusterInstance.SetupCluster(keyspace, []cluster.Shard{*shard0})
	if err != nil {
		return err
	}
	//Start MySql
	var mysqlCtlProcessList []*exec.Cmd
	for _, tablet := range shard0.Vttablets {
		log.Infof("Starting MySql for tablet %v", tablet.Alias)
		proc, err := tablet.MysqlctlProcess.StartProcess()
		if err != nil {
			return err
		}
		mysqlCtlProcessList = append(mysqlCtlProcessList, proc)
	}
	// Wait for mysql processes to start
	for _, proc := range mysqlCtlProcessList {
		err := proc.Wait()
		if err != nil {
			return err
		}
	}
	for _, tablet := range shard0.Vttablets {
		// Reset status, don't wait for the tablet status. We will check it later
		tablet.VttabletProcess.ServingStatus = ""
		// Start the tablet
		err := tablet.VttabletProcess.Setup()
		if err != nil {
			return err
		}
	}
	for _, tablet := range shard0.Vttablets {
		err := tablet.VttabletProcess.WaitForTabletTypes([]string{"SERVING", "NOT_SERVING"})
		if err != nil {
			return err
		}
	}
	return nil
}

// removeVttabletsFromTopology removes all the vttablets from the topology
func removeVttabletsFromTopology() error {
	for _, vttablet := range clusterInstance.Keyspaces[0].Shards[0].Vttablets {
		out, _ := clusterInstance.VtctlclientProcess.ExecuteCommandWithOutput("GetTablet", vttablet.Alias)
		log.Error("removeVttabletsFromTopology: ", out)

		err := clusterInstance.VtctlclientProcess.ExecuteCommand("DeleteTablet", "-allow_master", vttablet.Alias)
		if err != nil {
			return err
		}
	}
	clusterInstance.Keyspaces[0].Shards[0].Vttablets = nil
	return nil
}

// startVtorc is used to start the orchestrator with the given extra arguments
func startVtorc(t *testing.T, orcExtraArgs []string) {
	// Start vtorc
	clusterInstance.VtorcProcess = clusterInstance.NewOrcProcess(path.Join(os.Getenv("PWD"), "test_config.json"))
	clusterInstance.VtorcProcess.ExtraArgs = orcExtraArgs
	err := clusterInstance.VtorcProcess.Setup()
	require.NoError(t, err)
}

// stopVtorc is used to stop the orchestrator
func stopVtorc(t *testing.T) {
	// Stop vtorc
	if clusterInstance.VtorcProcess != nil {
		err := clusterInstance.VtorcProcess.TearDown()
		require.NoError(t, err)
	}
	clusterInstance.VtorcProcess = nil
}

// setupVttabletsAndVtorc is used to setup the vttablets and start the orchestrator
func setupVttabletsAndVtorc(t *testing.T, numReplicasReq int, numRdonlyReq int, orcExtraArgs []string) {
	// stop vtorc if it is running
	stopVtorc(t)

	// remove all the vttablets so that each test can add the amount that they require
	err := removeVttabletsFromTopology()
	require.NoError(t, err)

	for _, tablet := range replicaTablets {
		if numReplicasReq == 0 {
			break
		}
		err = cleanAndAddVttablet(t, tablet)
		require.NoError(t, err)
		numReplicasReq--
	}

	for _, tablet := range rdonlyTablets {
		if numRdonlyReq == 0 {
			break
		}
		err = cleanAndAddVttablet(t, tablet)
		require.NoError(t, err)
		numRdonlyReq--
	}

	if numRdonlyReq > 0 || numReplicasReq > 0 {
		t.Fatalf("more than available tablets requested. Please increase the constants numReplicas or numRdonly")
	}

	for _, tablet := range clusterInstance.Keyspaces[0].Shards[0].Vttablets {
		err = tablet.VttabletProcess.WaitForTabletTypes([]string{"SERVING", "NOT_SERVING"})
		require.NoError(t, err)
	}

	// start vtorc
	startVtorc(t, orcExtraArgs)
}

func cleanAndAddVttablet(t *testing.T, vttablet *cluster.Vttablet) error {
	// remove the database if it exists
	runSQL(t, "DROP DATABASE IF EXISTS vt_ks", vttablet, "")
	// reset the binlog
	runSQL(t, "RESET MASTER", vttablet, "")

	// add the vttablet to the topology
	err := clusterInstance.VtctlclientProcess.ExecuteCommand("InitTablet",
		"-port", fmt.Sprintf("%d", vttablet.VttabletProcess.Port),
		"-grpc_port", fmt.Sprintf("%d", vttablet.VttabletProcess.GrpcPort),
		"-hostname", vttablet.VttabletProcess.TabletHostname,
		"-mysql_port", fmt.Sprintf("%d", vttablet.MysqlctlProcess.MySQLPort),
		"-keyspace", vttablet.VttabletProcess.Keyspace,
		"-shard", vttablet.VttabletProcess.Shard,
		vttablet.Alias,
		vttablet.VttabletProcess.TabletType)
	if err != nil {
		return err
	}

	out, _ := clusterInstance.VtctlclientProcess.ExecuteCommandWithOutput("GetTablet", vttablet.Alias)
	log.Error("cleanAndAddVttablet: ", out)

	clusterInstance.Keyspaces[0].Shards[0].Vttablets = append(clusterInstance.Keyspaces[0].Shards[0].Vttablets, vttablet)
	return nil
}

func TestMain(m *testing.M) {
	exitcode, err := func() (int, error) {
		err := createClusterAndStartTopo()
		if err != nil {
			return 1, err
		}
		return m.Run(), nil
	}()

	cluster.PanicHandler(nil)
	clusterInstance.Teardown()
	killTablets(replicaTablets)
	killTablets(rdonlyTablets)

	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	} else {
		os.Exit(exitcode)
	}
}

// Cases to test:
// 1. create cluster with 1 replica and 1 rdonly, let orc choose master
// verify rdonly is not elected, only replica
// verify replication is setup
func TestMasterElection(t *testing.T) {
	defer cluster.PanicHandler(t)
	setupVttabletsAndVtorc(t, 1, 1, nil)
	keyspace := &clusterInstance.Keyspaces[0]
	shard0 := &keyspace.Shards[0]

	checkMasterTablet(t, clusterInstance, shard0.Vttablets[0])
	checkReplication(t, clusterInstance, shard0.Vttablets[0], shard0.Vttablets[1:])
}

// Cases to test:
// 1. create cluster with 1 replica and 1 rdonly, let orc choose master
// verify rdonly is not elected, only replica
// verify replication is setup
func TestSingleKeyspace(t *testing.T) {
	defer cluster.PanicHandler(t)
	setupVttabletsAndVtorc(t, 1, 1, []string{"-clusters_to_watch", "ks"})
	keyspace := &clusterInstance.Keyspaces[0]
	shard0 := &keyspace.Shards[0]

	checkMasterTablet(t, clusterInstance, shard0.Vttablets[0])
	checkReplication(t, clusterInstance, shard0.Vttablets[0], shard0.Vttablets[1:])
}

// Cases to test:
// 1. create cluster with 1 replica and 1 rdonly, let orc choose master
// verify rdonly is not elected, only replica
// verify replication is setup
func TestKeyspaceShard(t *testing.T) {
	defer cluster.PanicHandler(t)
	setupVttabletsAndVtorc(t, 1, 1, []string{"-clusters_to_watch", "ks/0"})
	keyspace := &clusterInstance.Keyspaces[0]
	shard0 := &keyspace.Shards[0]

	checkMasterTablet(t, clusterInstance, shard0.Vttablets[0])
	checkReplication(t, clusterInstance, shard0.Vttablets[0], shard0.Vttablets[1:])
}

// 2. bring down master, let orc promote replica
func TestDownMaster(t *testing.T) {
	defer cluster.PanicHandler(t)
	setupVttabletsAndVtorc(t, 2, 0, nil)
	keyspace := &clusterInstance.Keyspaces[0]
	shard0 := &keyspace.Shards[0]
	// find master from topo
	curMaster := shardMasterTablet(t, clusterInstance, keyspace, shard0)
	assert.NotNil(t, curMaster, "should have elected a master")

	// Make the current master database unavailable.
	err := curMaster.MysqlctlProcess.Stop()
	require.NoError(t, err)

	for _, tablet := range shard0.Vttablets {
		// we know we have only two tablets, so the "other" one must be the new master
		if tablet.Alias != curMaster.Alias {
			checkMasterTablet(t, clusterInstance, tablet)
			break
		}
	}
}

func waitForReadOnlyValue(t *testing.T, curMaster *cluster.Vttablet, expectValue int64) (match bool) {
	timeout := 15 * time.Second
	startTime := time.Now()
	for time.Since(startTime) < timeout {
		qr := runSQL(t, "select @@global.read_only as read_only", curMaster, "")
		require.NotNil(t, qr)
		row := qr.Named().Row()
		require.NotNil(t, row)
		readOnly, err := row.ToInt64("read_only")
		require.NoError(t, err)
		if readOnly == expectValue {
			return true
		}
		time.Sleep(time.Second)
	}
	return false
}

// 3. make master readonly, let orc repair
func TestMasterReadOnly(t *testing.T) {
	defer cluster.PanicHandler(t)
	setupVttabletsAndVtorc(t, 2, 0, nil)
	keyspace := &clusterInstance.Keyspaces[0]
	shard0 := &keyspace.Shards[0]

	// find master from topo
	curMaster := shardMasterTablet(t, clusterInstance, keyspace, shard0)
	assert.NotNil(t, curMaster, "should have elected a master")

	// Make the current master database read-only.
	runSQL(t, "set global read_only=ON", curMaster, "")

	// wait for repair
	match := waitForReadOnlyValue(t, curMaster, 0)
	require.True(t, match)
}

// 4. make replica ReadWrite, let orc repair
func TestReplicaReadWrite(t *testing.T) {
	defer cluster.PanicHandler(t)
	setupVttabletsAndVtorc(t, 2, 0, nil)
	keyspace := &clusterInstance.Keyspaces[0]
	shard0 := &keyspace.Shards[0]

	// find master from topo
	curMaster := shardMasterTablet(t, clusterInstance, keyspace, shard0)
	assert.NotNil(t, curMaster, "should have elected a master")

	var replica *cluster.Vttablet
	for _, tablet := range shard0.Vttablets {
		// we know we have only two tablets, so the "other" one must be the new master
		if tablet.Alias != curMaster.Alias {
			replica = tablet
			break
		}
	}
	// Make the replica database read-write.
	runSQL(t, "set global read_only=OFF", replica, "")

	// wait for repair
	match := waitForReadOnlyValue(t, replica, 1)
	require.True(t, match)
}

// 5. stop replication, let orc repair
func TestStopReplication(t *testing.T) {
	defer cluster.PanicHandler(t)
	setupVttabletsAndVtorc(t, 2, 0, nil)
	keyspace := &clusterInstance.Keyspaces[0]
	shard0 := &keyspace.Shards[0]

	// find master from topo
	curMaster := shardMasterTablet(t, clusterInstance, keyspace, shard0)
	assert.NotNil(t, curMaster, "should have elected a master")

	// TODO(deepthi): we should not need to do this, the DB should be created automatically
	_, err := curMaster.VttabletProcess.QueryTablet(fmt.Sprintf("create database IF NOT EXISTS vt_%s", keyspace.Name), keyspace.Name, false)
	require.NoError(t, err)

	var replica *cluster.Vttablet
	for _, tablet := range shard0.Vttablets {
		// we know we have only two tablets, so the "other" one must be the new master
		if tablet.Alias != curMaster.Alias {
			replica = tablet
			break
		}
	}
	require.NotNil(t, replica, "should be able to find a replica")
	// use vtctlclient to stop replication
	_, err = clusterInstance.VtctlclientProcess.ExecuteCommandWithOutput("StopReplication", replica.Alias)
	require.NoError(t, err)

	// wait for repair
	time.Sleep(15 * time.Second)
	// check replication is setup correctly
	checkReplication(t, clusterInstance, curMaster, []*cluster.Vttablet{replica})
}

// 6. setup replication from non-master, let orc repair
func TestReplicationFromOtherReplica(t *testing.T) {
	defer cluster.PanicHandler(t)
	setupVttabletsAndVtorc(t, 3, 0, nil)
	keyspace := &clusterInstance.Keyspaces[0]
	shard0 := &keyspace.Shards[0]

	// find master from topo
	curMaster := shardMasterTablet(t, clusterInstance, keyspace, shard0)
	assert.NotNil(t, curMaster, "should have elected a master")

	// TODO(deepthi): we should not need to do this, the DB should be created automatically
	_, err := curMaster.VttabletProcess.QueryTablet(fmt.Sprintf("create database IF NOT EXISTS vt_%s", keyspace.Name), keyspace.Name, false)
	require.NoError(t, err)

	var replica, otherReplica *cluster.Vttablet
	for _, tablet := range shard0.Vttablets {
		// we know we have only two tablets, so the "other" one must be the new master
		if tablet.Alias != curMaster.Alias {
			if replica == nil {
				replica = tablet
			} else {
				otherReplica = tablet
			}
		}
	}
	require.NotNil(t, replica, "should be able to find a replica")
	require.NotNil(t, otherReplica, "should be able to find 2nd replica")

	// point replica at otherReplica
	// Get master position
	hostname := "localhost"
	_, gtid := cluster.GetMasterPosition(t, *otherReplica, hostname)

	changeMasterCommand := fmt.Sprintf("STOP SLAVE; RESET MASTER; SET GLOBAL gtid_purged = '%s';"+
		"CHANGE MASTER TO MASTER_HOST='%s', MASTER_PORT=%d, MASTER_USER='vt_repl', MASTER_AUTO_POSITION = 1; START SLAVE", gtid, hostname, otherReplica.MySQLPort)
	result, err := clusterInstance.VtctlclientProcess.ExecuteCommandWithOutput("ExecuteFetchAsDba", replica.Alias, changeMasterCommand)
	require.NoError(t, err, result)

	// wait for repair
	time.Sleep(15 * time.Second)
	// check replication is setup correctly
	checkReplication(t, clusterInstance, curMaster, []*cluster.Vttablet{replica, otherReplica})
}

func TestRepairAfterTER(t *testing.T) {
	// test fails intermittently on CI, skip until it can be fixed.
	t.SkipNow()
	defer cluster.PanicHandler(t)
	setupVttabletsAndVtorc(t, 2, 0, nil)
	keyspace := &clusterInstance.Keyspaces[0]
	shard0 := &keyspace.Shards[0]

	// find master from topo
	curMaster := shardMasterTablet(t, clusterInstance, keyspace, shard0)
	assert.NotNil(t, curMaster, "should have elected a master")

	// TODO(deepthi): we should not need to do this, the DB should be created automatically
	_, err := curMaster.VttabletProcess.QueryTablet(fmt.Sprintf("create database IF NOT EXISTS vt_%s", keyspace.Name), keyspace.Name, false)
	require.NoError(t, err)

	var newMaster *cluster.Vttablet
	for _, tablet := range shard0.Vttablets {
		// we know we have only two tablets, so the "other" one must be the new master
		if tablet.Alias != curMaster.Alias {
			newMaster = tablet
			break
		}
	}

	// TER to other tablet
	_, err = clusterInstance.VtctlclientProcess.ExecuteCommandWithOutput("TabletExternallyReparented", newMaster.Alias)
	require.NoError(t, err)

	// wait for repair
	// TODO(deepthi): wait for condition instead of sleep
	time.Sleep(15 * time.Second)

	checkReplication(t, clusterInstance, newMaster, []*cluster.Vttablet{curMaster})
}

func shardMasterTablet(t *testing.T, cluster *cluster.LocalProcessCluster, keyspace *cluster.Keyspace, shard *cluster.Shard) *cluster.Vttablet {
	start := time.Now()
	for {
		now := time.Now()
		if now.Sub(start) > time.Second*60 {
			assert.FailNow(t, "failed to elect master before timeout")
		}
		result, err := cluster.VtctlclientProcess.ExecuteCommandWithOutput("GetShard", fmt.Sprintf("%s/%s", keyspace.Name, shard.Name))
		assert.Nil(t, err)

		var shardInfo topodatapb.Shard
		err = json2.Unmarshal([]byte(result), &shardInfo)
		assert.Nil(t, err)
		if shardInfo.MasterAlias == nil {
			log.Warningf("Shard %v/%v has no master yet, sleep for 1 second\n", keyspace.Name, shard.Name)
			time.Sleep(time.Second)
			continue
		}
		for _, tablet := range shard.Vttablets {
			if tablet.Alias == topoproto.TabletAliasString(shardInfo.MasterAlias) {
				return tablet
			}
		}
	}
}

// Makes sure the tablet type is master, and its health check agrees.
func checkMasterTablet(t *testing.T, cluster *cluster.LocalProcessCluster, tablet *cluster.Vttablet) {
	start := time.Now()
	for {
		now := time.Now()
		if now.Sub(start) > time.Second*60 {
			//log.Exitf("error")
			assert.FailNow(t, "failed to elect master before timeout")
		}
		result, err := cluster.VtctlclientProcess.ExecuteCommandWithOutput("GetTablet", tablet.Alias)
		require.NoError(t, err)
		var tabletInfo topodatapb.Tablet
		err = json2.Unmarshal([]byte(result), &tabletInfo)
		require.NoError(t, err)

		if topodatapb.TabletType_MASTER != tabletInfo.GetType() {
			log.Warningf("Tablet %v is not master yet, sleep for 1 second\n", tablet.Alias)
			time.Sleep(time.Second)
			continue
		} else {
			// allow time for tablet state to be updated after topo is updated
			time.Sleep(2 * time.Second)
			// make sure the health stream is updated
			result, err = cluster.VtctlclientProcess.ExecuteCommandWithOutput("VtTabletStreamHealth", "-count", "1", tablet.Alias)
			require.NoError(t, err)
			var streamHealthResponse querypb.StreamHealthResponse

			err = json2.Unmarshal([]byte(result), &streamHealthResponse)
			require.NoError(t, err)
			//if !streamHealthResponse.GetServing() {
			//	log.Exitf("stream health not updated")
			//}
			assert.True(t, streamHealthResponse.GetServing(), "stream health: %v", &streamHealthResponse)
			tabletType := streamHealthResponse.GetTarget().GetTabletType()
			require.Equal(t, topodatapb.TabletType_MASTER, tabletType)
			break
		}
	}
}

func checkReplication(t *testing.T, clusterInstance *cluster.LocalProcessCluster, master *cluster.Vttablet, replicas []*cluster.Vttablet) {
	validateTopology(t, clusterInstance, true)

	// create tables, insert data and make sure it is replicated correctly
	sqlSchema := `
		create table vt_insert_test (
		id bigint,
		msg varchar(64),
		primary key (id)
		) Engine=InnoDB
		`
	runSQL(t, sqlSchema, master, "vt_ks")
	confirmReplication(t, master, replicas)
}

func confirmReplication(t *testing.T, master *cluster.Vttablet, replicas []*cluster.Vttablet) {
	log.Infof("Insert data into master and check that it is replicated to replica")
	n := 2 // random value ...
	// insert data into the new master, check the connected replica work
	insertSQL := fmt.Sprintf("insert into vt_insert_test(id, msg) values (%d, 'test %d')", n, n)
	runSQL(t, insertSQL, master, "vt_ks")
	time.Sleep(100 * time.Millisecond)
	for _, tab := range replicas {
		err := checkInsertedValues(t, tab, n)
		require.NoError(t, err)
	}
}

func checkInsertedValues(t *testing.T, tablet *cluster.Vttablet, index int) error {
	// wait until it gets the data
	timeout := time.Now().Add(10 * time.Second)
	for time.Now().Before(timeout) {
		selectSQL := fmt.Sprintf("select msg from vt_insert_test where id=%d", index)
		qr := runSQL(t, selectSQL, tablet, "vt_ks")
		if len(qr.Rows) == 1 {
			return nil
		}
		time.Sleep(300 * time.Millisecond)
	}
	return fmt.Errorf("data is not yet replicated")
}

func validateTopology(t *testing.T, cluster *cluster.LocalProcessCluster, pingTablets bool) {
	if pingTablets {
		out, err := cluster.VtctlclientProcess.ExecuteCommandWithOutput("Validate", "-ping-tablets=true")
		require.NoError(t, err, out)
	} else {
		err := cluster.VtctlclientProcess.ExecuteCommand("Validate")
		require.NoError(t, err)
	}
}

func killTablets(vttablets []*cluster.Vttablet) {
	for _, tablet := range vttablets {
		log.Infof("Calling TearDown on tablet %v", tablet.Alias)
		_ = tablet.VttabletProcess.TearDown()
	}
}

func getMysqlConnParam(tablet *cluster.Vttablet, db string) mysql.ConnParams {
	connParams := mysql.ConnParams{
		Uname:      "vt_dba",
		UnixSocket: path.Join(os.Getenv("VTDATAROOT"), fmt.Sprintf("/vt_%010d/mysql.sock", tablet.TabletUID)),
	}
	if db != "" {
		connParams.DbName = db
	}
	return connParams
}

func runSQL(t *testing.T, sql string, tablet *cluster.Vttablet, db string) *sqltypes.Result {
	// Get Connection
	tabletParams := getMysqlConnParam(tablet, db)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := mysql.Connect(ctx, &tabletParams)
	require.Nil(t, err)
	defer conn.Close()

	// runSQL
	return execute(t, conn, sql)
}

func execute(t *testing.T, conn *mysql.Conn, query string) *sqltypes.Result {
	t.Helper()
	qr, err := conn.ExecuteFetch(query, 1000, true)
	require.Nil(t, err)
	return qr
}

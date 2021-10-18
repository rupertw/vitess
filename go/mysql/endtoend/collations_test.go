package endtoend

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/text/encoding/unicode/utf32"

	"vitess.io/vitess/go/mysql"
	"vitess.io/vitess/go/mysql/collations"
	"vitess.io/vitess/go/sqltypes"
	"vitess.io/vitess/go/vt/sqlparser"
)

func getSQLQueries(t *testing.T, testfile string) []string {
	tf, err := os.Open(testfile)
	if err != nil {
		t.Fatal(err)
	}
	defer tf.Close()

	var chunks []string
	var curchunk bytes.Buffer

	addchunk := func() {
		if curchunk.Len() > 0 {
			stmts, err := sqlparser.SplitStatementToPieces(curchunk.String())
			if err != nil {
				t.Fatal(err)
			}
			chunks = append(chunks, stmts...)
			curchunk.Reset()
		}
	}

	scanner := bufio.NewScanner(tf)
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "--") {
			addchunk()
			chunks = append(chunks, scanner.Text())
		} else {
			if curchunk.Len() > 0 {
				curchunk.WriteByte(' ')
			}
			curchunk.Write(scanner.Bytes())
		}
	}
	addchunk()
	return chunks
}

type TestOnResults interface {
	Test(t *testing.T, result *sqltypes.Result)
}

type uca900CollationTest struct {
	collation string
}

var defaultUtf32 = utf32.UTF32(utf32.BigEndian, utf32.IgnoreBOM)

func parseUtf32cp(b []byte) []byte {
	var hexbuf [16]byte
	c, err := hex.Decode(hexbuf[:], b)
	if err != nil {
		return nil
	}
	utf8, _ := defaultUtf32.NewDecoder().Bytes(hexbuf[:c])
	return utf8
}

func parseWeightString(b []byte) []byte {
	dst := make([]byte, hex.DecodedLen(len(b)))
	n, err := hex.Decode(dst, b)
	if err != nil {
		return nil
	}
	return dst[:n]
}

func (u *uca900CollationTest) Test(t *testing.T, result *sqltypes.Result) {
	coll := collations.LookupByName(u.collation)
	if coll == nil {
		t.Fatalf("unknown collation %q", u.collation)
	}

	var checked, errors int
	for _, row := range result.Rows {
		if row[1].Len() == 0 {
			continue
		}
		utf8Input := parseUtf32cp(row[0].ToBytes())
		if utf8Input == nil {
			t.Errorf("[%s] failed to parse UTF32-encoded codepoint: %s (%s)", u.collation, row[0], row[2].ToString())
			errors++
			continue
		}

		expectedWeightString := parseWeightString(row[1].ToBytes())
		if expectedWeightString == nil {
			t.Errorf("[%s] failed to parse weight string: %s (%s)", u.collation, row[1], row[2].ToString())
			errors++
			continue
		}

		weightString := coll.WeightString(make([]byte, 0, 128), utf8Input, 0)
		if !bytes.Equal(weightString, expectedWeightString) {
			t.Errorf("[%s] mismatch for %s (%v): \n\twant: %v\n\tgot:  %v", u.collation, row[2].ToString(), utf8Input, expectedWeightString, weightString)
			errors++
		}
		checked++
	}

	t.Logf("uca900CollationTest[%s]: checked %d codepoints, %d failed (%.02f%%)", u.collation, checked, errors, float64(errors)/float64(checked)*100.0)
}

func processSQLTest(t *testing.T, testfile string, conn *mysql.Conn) {
	var curtest TestOnResults

	for _, query := range getSQLQueries(t, testfile) {
		if strings.HasPrefix(query, "--") {
			switch {
			case strings.HasPrefix(query, "--source "):
				include := strings.TrimPrefix(query, "--source ")
				include = path.Join("testdata/mysqltest", include)
				processSQLTest(t, include, conn)

			case strings.HasPrefix(query, "--test:uca0900 "):
				collation := strings.TrimPrefix(query, "--test:uca0900 ")
				curtest = &uca900CollationTest{collation}

			case query == "--disable_warnings" || query == "--enable_warnings":
			case query == "--disable_query_log" || query == "--enable_query_log":

			default:
				t.Logf("unsupported statement: %q", query)
			}
			continue
		}

		res, err := conn.ExecuteFetch(query, -1, false)
		if err != nil {
			t.Fatalf("failed to execute %q: %v", query, err)
		}

		if curtest != nil {
			curtest.Test(t, res)
			curtest = nil
		}
	}
}

var testOneCollation = flag.String("test-one-collation", "", "")

func TestCollationsOnMysqld(t *testing.T) {
	conn, err := mysql.Connect(context.Background(), &connParams)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if !strings.HasPrefix(conn.ServerVersion, "8.0.") {
		t.Skipf("uca900 collations are only supported in MySQL 8.0+")
	}

	if *testOneCollation != "" {
		processSQLTest(t, fmt.Sprintf("testdata/mysqltest/suite/collations/%s.test", *testOneCollation), conn)
		return
	}

	testfiles, _ := filepath.Glob("testdata/mysqltest/suite/collations/*.test")
	for _, testfile := range testfiles {
		t.Run(testfile, func(t *testing.T) {
			processSQLTest(t, testfile, conn)
		})
	}
}

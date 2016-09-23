package dbutil

import (
	"errors"
	"runtime"
	"testing"
)

func TestConnect(t *testing.T) {
	_, err := Connect("10.11.12.13", 3306, "kchu", "test")
	checkFatal(t, err)
}

func TestDatabaseList(t *testing.T) {
	// dbh, err := Connect("10.11.12.13", 3306, "kchu", "test")
	// dbh.Exec("CREATE DATABASE dbutil_test1")
	// dbh.Exec("CREATE DATABASE dbutil_test2")
	// dbh.Exec("CREATE DATABASE dbutil_test3")

	// dbs, err := DatabaseList(dbh)
	// 8888gTcheckFatal(t, err)

	// for i, db := range dbs {
	// 	expected := "dbutil_test" + strconv.Itoa(i)
	// 	if db != expected {
	// 		checkFatal(t, fmt.Errorf("Expected"))
	// 	}
	// }
	// t.Log(dbs)

	// dbh.Exec("DROP DATABASE dbutil_test1")
	// dbh.Exec("DROP DATABASE dbutil_test2")
	// dbh.Exec("DROP DATABASE dbutil_test3")
}

func TestDrop(t *testing.T) {
	dbh, err := Connect("10.11.12.13", 3306, "kchu", "test")
	_, err = dbh.Exec("CREATE DATABASE dbutil_test1")
	checkFatal(t, err)

	_, err = dbh.Exec("USE dbutil_test1")
	checkFatal(t, err)

	err = Drop(dbh, []string{"dbutil_test1"})

	_, err = dbh.Exec("USE dbutil_test1")
	if err == nil {
		checkFatal(t, errors.New("USE query should have thrown the erro since dbutil_test1 is dropped"))
	}
}

func checkFatal(t *testing.T, err error) {
	if err == nil {
		return
	}

	// The failure happens at wherever we were called, not here
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		t.Fatalf("Unable to get caller")
	}
	t.Fatalf("Fail at %v:%v; %v", file, line, err)
}

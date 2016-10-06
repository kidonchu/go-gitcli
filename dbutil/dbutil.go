package dbutil

import (
	"database/sql"
	"fmt"
	"regexp"

	_ "github.com/go-sql-driver/mysql" // mysql driver
)

// Connect returns db instance of given db information
func Connect(host string, port int32, user string, pass string) (*sql.DB, error) {

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/hosted", user, pass, host, port)
	db, err := sql.Open("mysql", dsn)

	if err != nil {
		return nil, fmt.Errorf("Unable to connect to database with `%s`", dsn)
	}

	return db, nil
}

// FindDbs returns an array of pattern-matched dbs in the db connection
func FindDbs(dbh *sql.DB, pattern string) ([]string, error) {

	var dbs []string

	rows, err := dbh.Query("SHOW DATABASES")
	if err != nil {
		return nil, fmt.Errorf("Unable to get a list of databases: %+v", err)
	}
	defer rows.Close()

	regex, _ := regexp.Compile(pattern)
	for rows.Next() {
		var database string
		if err := rows.Scan(&database); err != nil {
			return nil, fmt.Errorf("Database could not be fetched from result rows: %+v", err)
		}
		matched := regex.FindString(database)
		if matched != "" {
			dbs = append(dbs, database)
		}
	}

	return dbs, nil
}

// Drop drops provided dbs
func Drop(dbh *sql.DB, dbs []string) error {
	for _, db := range dbs {
		fmt.Printf("Deleting database: `%s`...\n", db)
		_, err := dbh.Exec(fmt.Sprintf("DROP DATABASE `%s`", db))
		if err != nil {
			return fmt.Errorf("Error while deleting database `%s`: %+v", db, err)
		}
	}
	return nil
}

package main

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/guardian/fastly-log-xtractor/models"
	_ "github.com/mattn/go-sqlite3"
)

const sqliteTableName = "fastly_logs"

var sqliteColumnTypes = map[string]string{
	"HTTP Response Code": "INTEGER",
	"Response body size": "INTEGER",
	"Is on edge?":        "BOOLEAN",
}

var (
	nonAlnumRe   = regexp.MustCompile(`[^a-z0-9\s]+`)
	whitespaceRe = regexp.MustCompile(`\s+`)
)

func normaliseColumnName(name string) string {
	name = nonAlnumRe.ReplaceAllString(strings.ToLower(name), "")
	return whitespaceRe.ReplaceAllString(strings.TrimSpace(name), "_")
}

func sqliteSchema() (createStmt string, insertStmt string) {
	headers := models.CSVHeaderRow()
	colDefs := make([]string, len(headers))
	colNames := make([]string, len(headers))
	placeholders := make([]string, len(headers))
	for i, h := range headers {
		colType, ok := sqliteColumnTypes[h]
		if !ok {
			colType = "TEXT"
		}
		colNames[i] = normaliseColumnName(h)
		colDefs[i] = colNames[i] + " " + colType
		placeholders[i] = "?"
	}
	createStmt = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", sqliteTableName, strings.Join(colDefs, ", "))
	insertStmt = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", sqliteTableName, strings.Join(colNames, ", "), strings.Join(placeholders, ", "))
	return
}

// Argument order must match models.CSVHeaderRow()
func sqliteRowValues(l *models.FastlyLog) []interface{} {
	return []interface{}{
		l.Timestamp,
		l.ClientIp,
		l.GeoCountry,
		l.GeoCity,
		l.RequestMethod,
		l.ResponseStatus,
		l.RequestProtocol,
		l.RequestReferrer,
		l.RequestUserAgent,
		l.ResponseState,
		l.ResponseReason,
		l.ResponseBodySize,
		l.FastlyServer,
		l.FastlyIsEdge,
		l.Host,
		l.Url,
	}
}

func AsyncSQLiteWriter(filename *string, contentCh chan models.FastlyLog, commitEvery int) chan error {
	errCh := make(chan error, 1)

	fail := func(err error) {
		log.Printf("ERROR Could not write %s: %s", *filename, err)
		errCh <- err
		close(errCh)
	}

	go func() {
		if commitEvery < 1 {
			fail(fmt.Errorf("commitEvery must be at least 1, got %d", commitEvery))
			return
		}

		db, err := sql.Open("sqlite3", *filename)
		if err != nil {
			fail(err)
			return
		}
		defer db.Close()

		createStmt, insertStmt := sqliteSchema()
		if _, err = db.Exec(createStmt); err != nil {
			fail(err)
			return
		}

		stmt, err := db.Prepare(insertStmt)
		if err != nil {
			fail(err)
			return
		}
		defer stmt.Close()

		// Batching rows into transactions is far faster than autocommitting each row
		tx, err := db.Begin()
		if err != nil {
			fail(err)
			return
		}
		txStmt := tx.Stmt(stmt)
		count := 0

		for {
			rec, haveMore := <-contentCh
			if !haveMore {
				if err = tx.Commit(); err != nil {
					fail(err)
					return
				}
				log.Print("All records now written")
				close(errCh)
				return
			}

			if _, err = txStmt.Exec(sqliteRowValues(&rec)...); err != nil {
				tx.Rollback()
				fail(err)
				return
			}

			count++
			if count >= commitEvery {
				if err = tx.Commit(); err != nil {
					fail(err)
					return
				}
				if tx, err = db.Begin(); err != nil {
					fail(err)
					return
				}
				txStmt = tx.Stmt(stmt)
				count = 0
			}
		}
	}()
	return errCh
}

package lgpn

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
)

var schema = `CREATE TABLE IF NOT EXISTS names (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	not_before INTEGER NOT NULL,
	not_after INTEGER NOT NULL
)`

type cache struct {
	db *sql.DB
}

// newCache opens the existing cache or, if necessary, creates a new empty
// cache.
func newCache(ctx context.Context, cacheDir string) (*cache, error) {
	path := filepath.Join(cacheDir, "lgpn.sqlite3")
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("unable to open cache db: %w", err)
	}

	_, err = db.ExecContext(ctx, schema)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("unable to create schema: %w", err)
	}

	return &cache{db: db}, nil
}

// Close closes the underlying database.
func (c *cache) Close() {
	c.db.Close()
}

// Count returns the number of names in the cache.
func (c *cache) Count(ctx context.Context) (int, error) {
	row := c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM names")

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Populate reads rows from the channel and inserts them into the database.
func (c *cache) Populate(ctx context.Context, rows <-chan nameRow) error {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("unable to start transaction: %w", err)
	}
	defer tx.Rollback() // no-op after Commit()

	stmt, err := tx.Prepare(`INSERT INTO names (name, not_before, not_after) VALUES (?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("unable to prepare insert query: %w", err)
	}
	defer stmt.Close()

	for row := range rows {
		_, err = stmt.Exec(row.Name, row.NotBefore, row.NotAfter)
		if err != nil {
			return fmt.Errorf("insert failed for row %v: %w", row, err)
		}
	}

	return tx.Commit()
}

// Names returns a channel which emits every name in the cache. See the
// package-level Names function for details about the channel mechanics.
func (c *cache) Names(ctx context.Context) (<-chan string, error) {
	rows, err := c.db.QueryContext(ctx, "SELECT name FROM names")
	if err != nil {
		return nil, fmt.Errorf("unable to query names: %w", err)
	}

	names := make(chan string)

	go func() {
		defer close(names)
		var err error
		done := ctx.Done()

		for rows.Next() {
			var name string
			err = rows.Scan(&name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "unable to read row: %v", err)
				return
			}

			select {
			case <-done:
				return
			case names <- name:
			}
		}
	}()

	return names, nil
}

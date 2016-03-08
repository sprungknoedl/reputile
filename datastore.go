package main

import (
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"

	"golang.org/x/net/context"
)

type Datastore struct {
	*sql.DB
	create *sql.Stmt
}

func NewDatastore(url string) (*Datastore, error) {
	conn, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	store := &Datastore{conn, nil}
	store.create, err = conn.Prepare(`
		INSERT INTO entries
			(key, source, domain, ip4, last, category, description)
		VALUES
			($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (key) DO UPDATE SET
			last = $5, category = $6, description = $7`)
	if err != nil {
		return nil, err
	}

	return store, nil
}

type Entry struct {
	Source      string
	Domain      string
	IP4         string
	Last        time.Time
	Category    string
	Description string

	Err error
}

func SendError(err error) *Entry {
	return &Entry{Err: err}
}

func (e Entry) Key() string {
	return fmt.Sprintf("%s|%s|%s", e.Source, e.Domain, e.IP4)
}

func Store(ctx context.Context, e *Entry) error {
	db := ctx.Value(databaseKey).(*Datastore)
	_, err := db.create.Exec(
		e.Key(),
		e.Source,
		e.Domain,
		e.IP4,
		time.Now(),
		e.Category,
		e.Description)
	return err
}

var (
	psql             = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	queryTranslation = map[string]string{
		"source":      "source = ?",
		"domain":      "domain = ?",
		"ip4":         "ip4 = ?",
		"last":        "last > ?",
		"category":    "category = ?",
		"description": "description = ?",
	}
)

func Find(ctx context.Context, query map[string]string) chan *Entry {
	ch := make(chan *Entry)
	db := ctx.Value(databaseKey).(*Datastore)

	go func() {
		defer close(ch)

		builder := psql.
			Select("source", "domain", "ip4", "last", "category", "description").
			From("entries").
			OrderBy("source", "domain", "ip4")

		for key, value := range query {
			if pred, ok := queryTranslation[key]; ok {
				builder = builder.Where(pred, value)
			}
		}

		rows, err := builder.RunWith(db.DB).Query()
		if err != nil {
			ch <- SendError(err)
			return
		}

		defer rows.Close()
		for rows.Next() {
			e := &Entry{}
			err := rows.Scan(
				&e.Source,
				&e.Domain,
				&e.IP4,
				&e.Last,
				&e.Category,
				&e.Description)
			if err != nil {
				ch <- SendError(err)
				return
			}

			select {
			case <-ctx.Done():
				ch <- SendError(ctx.Err())
				break
			case ch <- e:
			}
		}
	}()

	return ch
}

func CountEntries(ctx context.Context) (int, error) {
	return getCount(ctx, `SELECT COUNT(*) FROM entries`)
}

func CountSources(ctx context.Context) (int, error) {
	return getCount(ctx, `SELECT COUNT(DISTINCT(source)) FROM entries`)
}

func getCount(ctx context.Context, query string) (int, error) {
	db := ctx.Value(databaseKey).(*Datastore)

	count := 0
	rows, err := db.Query(query)
	if err != nil {
		return 0, nil
	}

	// scan first row, queries should only return one but we are not
	// enforcing this
	rows.Next()
	err = rows.Scan(&count)
	rows.Close()

	return count, err
}

package main

import (
	"context"
	"database/sql"
	"net"
	"time"

	"github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
	"github.com/sprungknoedl/reputile/lists"
)

var (
	psql             = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	queryTranslation = map[string]string{
		"source":      "source = ?",
		"domain":      "domain = ?",
		"ip":          "ip <<= ?",
		"last":        "last > ?",
		"category":    "category = ?",
		"description": "description = ?",
	}
)

type Datastore struct {
	db *sql.DB
}

func NewDatastore(url string) (*Datastore, error) {
	conn, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	_, err = conn.Exec(`CREATE TABLE IF NOT EXISTS entries (
		key text NOT NULL PRIMARY KEY,
		source text NOT NULL,
		domain text NOT NULL,
		ip inet,
		category text NOT NULL,
		description text NOT NULL,
		last timestamp with time zone NOT NULL
		);`)
	if err != nil {
		return nil, err
	}

	return &Datastore{db: conn}, nil
}

func (db *Datastore) Store(ctx context.Context, e *lists.Entry) error {
	_, err := db.db.ExecContext(ctx, `
		INSERT INTO entries
		(key, source, domain, ip, last, category, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (key) DO UPDATE SET
		last = $5, category = $6, description = $7`,
		e.Key(),
		e.Source,
		e.Domain,
		sql.NullString{Valid: e.IP != nil, String: e.IP.String()},
		time.Now(),
		e.Category,
		e.Description)
	return err
}

func (db *Datastore) Prune(ctx context.Context) error {
	_, err := db.db.ExecContext(ctx, `
		DELETE FROM entries
		WHERE last < (now() - interval '7d')`)
	return err
}

func (db *Datastore) Find(ctx context.Context, query map[string]string) chan *lists.Entry {
	ch := make(chan *lists.Entry)

	go func() {
		defer close(ch)

		builder := psql.
			Select("source", "domain", "ip", "last", "category", "description").
			From("entries").
			OrderBy("source", "domain", "ip")

		for key, value := range query {
			if pred, ok := queryTranslation[key]; ok {
				builder = builder.Where(pred, value)
			}
		}

		query, args, err := builder.ToSql()
		if err != nil {
			ch <- lists.SendError(err)
			return
		}

		rows, err := db.db.QueryContext(ctx, query, args...)
		if err != nil {
			ch <- lists.SendError(err)
			return
		}

		defer rows.Close()
		for rows.Next() {
			e := &lists.Entry{}
			ip := sql.NullString{}
			err := rows.Scan(
				&e.Source,
				&e.Domain,
				&ip,
				&e.Last,
				&e.Category,
				&e.Description)

			if err != nil {
				ch <- lists.SendError(err)
				return
			}

			if ip.Valid {
				e.IP = net.ParseIP(ip.String)
			}

			select {
			case <-ctx.Done():
				ch <- lists.SendError(ctx.Err())
				break
			case ch <- e:
			}
		}
	}()

	return ch
}

func (db *Datastore) CountEntries(ctx context.Context) (int, error) {
	count := 0
	err := db.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM entries`).
		Scan(&count)
	if err != nil {
		return 0, nil
	}

	return count, nil
}

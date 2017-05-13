package main

import (
	"context"
	"database/sql"
	"net"
	"time"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
	"github.com/sprungknoedl/reputile/model"
)

type Datastore struct {
	*sql.DB
	create *sql.Stmt
	prune  *sql.Stmt
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

	store := &Datastore{conn, nil, nil}
	store.create, err = conn.Prepare(`
		INSERT INTO entries
			(key, source, domain, ip, last, category, description)
		VALUES
			($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (key) DO UPDATE SET
			last = $5, category = $6, description = $7`)
	if err != nil {
		return nil, err
	}

	store.prune, err = conn.Prepare(`
		DELETE FROM entries
		WHERE last < (now() - interval '7d')`)
	if err != nil {
		return nil, err
	}

	return store, nil
}

func (db *Datastore) Store(ctx context.Context, e *model.Entry) error {
	_, err := db.create.Exec(
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
	_, err := db.prune.Exec()
	return err
}

func (db *Datastore) Find(ctx context.Context, query map[string]string) chan *model.Entry {
	ch := make(chan *model.Entry)

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

		rows, err := builder.RunWith(db.DB).Query()
		if err != nil {
			ch <- model.SendError(err)
			return
		}

		defer rows.Close()
		for rows.Next() {
			e := &model.Entry{}
			ip := sql.NullString{}
			err := rows.Scan(
				&e.Source,
				&e.Domain,
				&ip,
				&e.Last,
				&e.Category,
				&e.Description)

			if err != nil {
				ch <- model.SendError(err)
				return
			}

			if ip.Valid {
				e.IP = net.ParseIP(ip.String)
			}

			select {
			case <-ctx.Done():
				ch <- model.SendError(ctx.Err())
				break
			case ch <- e:
			}
		}
	}()

	return ch
}

func (db *Datastore) CountEntries(ctx context.Context) (int, error) {
	count := 0
	err := db.QueryRow(`SELECT COUNT(*) FROM entries`).Scan(&count)
	if err != nil {
		return 0, nil
	}

	return count, nil
}

var (
	psql             = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	queryTranslation = map[string]string{
		"source":      "source = ?",
		"domain":      "domain = ?",
		"ip":          "ip <<= ?",
		"last":        "last > ?",
		"category":    "category = ?",
		"description": "description = ?",
	}
)

package main

import (
	"fmt"
	"time"

	"github.com/jackc/pgx"

	"golang.org/x/net/context"
)

type Datastore struct {
	*pgx.Conn
}

func NewDatastore(url string) (*Datastore, error) {
	cfg, err := pgx.ParseURI(url)
	if err != nil {
		return nil, err
	}

	conn, err := pgx.Connect(cfg)
	if err != nil {
		return nil, err
	}

	_, err = conn.Prepare("create-entry", `
		INSERT INTO entries
			(key, source, domain, ip4, last, category, description)
		VALUES
			($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (key) DO UPDATE SET
			last = $5, category = $6, description = $7`)
	if err != nil {
		return nil, err
	}

	_, err = conn.Prepare("read-entry", `
	SELECT 
		source, domain, ip4, last, category, description
	FROM
		entries
	ORDER BY
		source DESC,
		domain DESC,
		ip4 DESC`)
	if err != nil {
		return nil, err
	}

	return &Datastore{conn}, nil
}

type Entry struct {
	Source      string    `db:"source"`
	Domain      string    `db:"domain"`
	IP4         string    `db:"ip4"`
	Last        time.Time `db:"last"`
	Category    string    `db:"category"`
	Description string    `db:"description"`

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
	_, err := db.Exec("create-entry",
		e.Key(),
		e.Source,
		e.Domain,
		e.IP4,
		time.Now(),
		e.Category,
		e.Description)
	return err
}

func Find(ctx context.Context, query map[string]string) chan *Entry {
	ch := make(chan *Entry)
	db := ctx.Value(databaseKey).(*Datastore)

	go func() {
		defer close(ch)
		rows, err := db.Query("read-entry")
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

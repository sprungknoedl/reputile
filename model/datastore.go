package model

import (
	"database/sql"
	"fmt"
	"net"
	"time"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
	"github.com/sprungknoedl/reputile/lib"

	"golang.org/x/net/context"
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

type Entry struct {
	Source      string
	Domain      string
	IP          net.IP
	Last        time.Time
	Category    string
	Description string

	Err error
}

func SendError(err error) *Entry {
	return &Entry{Err: err}
}

func (e Entry) Key() string {
	return fmt.Sprintf("%s|%s|%s", e.Source, e.Domain, e.IP)
}

func Store(ctx context.Context, e *Entry) error {
	db := ctx.Value(lib.DatabaseKey).(*Datastore)
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

func Prune(ctx context.Context) error {
	db := ctx.Value(lib.DatabaseKey).(*Datastore)
	_, err := db.prune.Exec()
	return err
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

func Find(ctx context.Context, query map[string]string) chan *Entry {
	ch := make(chan *Entry)
	db := ctx.Value(lib.DatabaseKey).(*Datastore)

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
			ch <- SendError(err)
			return
		}

		defer rows.Close()
		for rows.Next() {
			e := &Entry{}
			ip := sql.NullString{}
			err := rows.Scan(
				&e.Source,
				&e.Domain,
				&ip,
				&e.Last,
				&e.Category,
				&e.Description)

			if err != nil {
				ch <- SendError(err)
				return
			}

			if ip.Valid {
				e.IP = net.ParseIP(ip.String)
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
	db := ctx.Value(lib.DatabaseKey).(*Datastore)

	count := 0
	err := db.QueryRow(`SELECT COUNT(*) FROM entries`).Scan(&count)
	if err != nil {
		return 0, nil
	}

	return count, nil
}

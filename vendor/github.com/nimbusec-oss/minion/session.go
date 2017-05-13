package minion

import (
	"net/http"

	"github.com/gorilla/sessions"
)

// Default retrieves a value from the given session. If the value does not
// exist in the session, a provided default is returned
func Default(session *sessions.Session, name string, def interface{}) interface{} {
	value, ok := session.Values[name]
	if !ok {
		return def
	}

	return value
}

func RenewSessionID(w http.ResponseWriter, r *http.Request, session *sessions.Session) *sessions.Session {
	age := session.Options.MaxAge

	// invalidate old session
	session.Options.MaxAge = -1
	session.Save(r, w)

	// force new session id
	session.ID = ""
	session.Options.MaxAge = age
	return session
}

type NilStore struct{}

func (store NilStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return store.New(r, name)
}

func (store NilStore) New(r *http.Request, name string) (*sessions.Session, error) {
	return &sessions.Session{IsNew: true}, nil
}

func (store NilStore) Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error {
	return nil
}

package minion

import "net/http"

// Principal is an entity that can be authenticated and verified.
type Principal interface {
	Authenticated() bool
	ID() string
	HasAnyRole(roles ...string) bool
}

// Anonymous implements the Principal interface for unauthenticated
// users and can be used as a fallback principal when none is set
// in the current session.
type Anonymous struct{}

// Authenticated returns always false, because Anonymous users are not
// authenticated.
func (a Anonymous) Authenticated() bool { return false }

// ID retunrs always the string `anonymous` as ID for unauthenticated
// users.
func (a Anonymous) ID() string { return "anonymous" }

// HasAnyRole returns always false for any role, because Anonymous users
// are not authenticated.
func (a Anonymous) HasAnyRole(roles ...string) bool { return false }

// GetPrincipal returns the principal from the current session.
func (m *Minion) GetPrincipal(r *http.Request) Principal {
	session, _ := m.Sessions.Get(r, m.SessionName)
	principal := Default(session, PrincipalKey, Anonymous{}).(Principal)
	return principal
}

// Secured requires that the user has at least one of the provided roles before
// the request is forwarded to the secured handler.
func (m *Minion) Secured(fn http.HandlerFunc, roles ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal := m.GetPrincipal(r)
		if !principal.Authenticated() {
			m.Unauthorized(w, r)
			return
		}

		if !principal.HasAnyRole(roles...) {
			m.Forbidden(w, r)
			return
		}

		fn(w, r)
	}
}

// defaultUnauthorizedHandler is the default handler for minion.Unauthorized
func (m *Minion) defaultUnauthorizedHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := m.Sessions.Get(r, m.SessionName)
	session.Values[RedirectKey] = r.URL.String()
	session.Save(r, w)

	http.Redirect(w, r, m.UnauthorizedURL, http.StatusSeeOther)
}

// defaultForbiddenHandler is the default handler for minion.Forbidden
func (m *Minion) defaultForbiddenHandler(w http.ResponseWriter, r *http.Request) {
	m.HTML(w, r, http.StatusForbidden, m.ForbiddenTemplate, V{})
}

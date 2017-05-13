package minion

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/sessions"
)

func init() {
	gob.Register(Anonymous{})
}

const (
	// PrincipalKey is the key used for the principal in the user session.
	PrincipalKey = "_principal"

	// RedirectKey is the key used for the original URL before redirecting to the login site.
	RedirectKey = "_redirect"
)

// Logger is a simple interface to describe something that can write log output.
type Logger interface {
	Printf(fmt string, v ...interface{})
}

// Option is a functional configuration type that can be used to tailor
// the Minion instance during creation.
type Option func(*Minion) *Minion

// Minion implements basic building blocks that most http servers require
type Minion struct {
	Debug  bool
	Logger Logger

	// Unauthorized is called for Secured handlers where no authenticated
	// principal is found in the current session. The default handler will
	// redirect the user to `UnauthorizedURL` and store the original URL
	// in the session.
	Unauthorized    func(w http.ResponseWriter, r *http.Request)
	UnauthorizedURL string

	// Forbidden is called for Secured handlers where an authenticated principal
	// does not have enough permission to view the resource. The default handler
	// will execute the HTML template `ForbiddenTemplate`.
	Forbidden         func(w http.ResponseWriter, r *http.Request)
	ForbiddenTemplate string

	// Error is called for any error that occur during the request processing, be
	// it client side errors (4xx status) or server side errors (5xx status). The
	// default handler will execute the HTML template `ErrorTemplate`.
	Error         func(w http.ResponseWriter, r *http.Request, code int, err error)
	ErrorTemplate string

	Sessions    sessions.Store
	SessionName string

	Templates       *template.Template
	TemplateFuncMap template.FuncMap
}

// NewMinion creates a new minion instance.
func NewMinion(options ...Option) *Minion {
	m := &Minion{
		Debug:  os.Getenv("DEBUG") == "true",
		Logger: log.New(os.Stderr, "", log.LstdFlags),

		UnauthorizedURL:   "/login",
		ErrorTemplate:     "500.html",
		ForbiddenTemplate: "403.html",

		Sessions: NilStore{},
		TemplateFuncMap: template.FuncMap{
			"div": func(dividend, divisor int) float64 {
				return float64(dividend) / float64(divisor)
			},
			"json": func(v interface{}) template.JS {
				b, _ := json.MarshalIndent(v, "", "  ")
				return template.JS(b) // nolint: gas
			},
			"dict": func(values ...interface{}) (map[string]interface{}, error) {
				if len(values)%2 != 0 {
					return nil, errors.New("invalid dict call")
				}
				dict := make(map[string]interface{}, len(values)/2)
				for i := 0; i < len(values); i += 2 {
					key, ok := values[i].(string)
					if !ok {
						return nil, errors.New("dict keys must be strings")
					}
					dict[key] = values[i+1]
				}
				return dict, nil
			},
		},
	}

	// default handlers
	m.Unauthorized = m.defaultUnauthorizedHandler
	m.Forbidden = m.defaultForbiddenHandler
	m.Error = m.defaultErrorHandler

	// apply functional configuration
	for _, option := range options {
		m = option(m)
	}

	return m
}

// defaultErrorHandler is the default handler for minion.Error
func (m *Minion) defaultErrorHandler(w http.ResponseWriter, r *http.Request, code int, err error) {
	m.Logger.Printf("error: %v", err)
	m.HTML(w, r, code, m.ErrorTemplate, V{
		"code":  code,
		"error": err.Error(),
	})
}

// JSON outputs the data encoded as JSON.
func (m *Minion) JSON(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	err := JSON(w, r, code, data)
	if err != nil {
		m.Logger.Printf("failed to encode json: %v", err)
	}
}

// HTML outputs a rendered HTML template to the client.
func (m *Minion) HTML(w http.ResponseWriter, r *http.Request, code int, name string, data V) {
	// reload templates in debug mode
	if m.Templates == nil || m.Debug {
		err := m.LoadTemplates()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			m.Logger.Printf("failed to parse templates: %v", err)
			return
		}
	}

	w.Header().Add("content-type", "text/html; charset=utf-8")
	w.WriteHeader(code)

	err := m.Templates.ExecuteTemplate(w, name, data)
	if err != nil {
		m.Logger.Printf("failed to execute template %q: %v", name, err)
		return
	}
}

// LoadTemplates loads the html/template files from the filesystem.
func (m *Minion) LoadTemplates() error {
	m.Templates = template.New("").Funcs(m.TemplateFuncMap)
	err := filepath.Walk("./templates", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			b, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			m.Templates, err = m.Templates.
				New(strings.TrimPrefix(path, "templates/")).
				Parse(string(b))
			return err
		}
		return nil
	})

	return err
}

// BindingResult holds validation errors of the binding process from a HTML
// form to a Go struct.
type BindingResult map[string]string

// Valid returns whether the binding was successfull or not.
func (br BindingResult) Valid() bool {
	return len(br) == 0
}

// Fail marks the binding as failed and stores an error for the given field
// that caused the form binding to fail.
func (br BindingResult) Fail(field, err string) {
	br[field] = err
}

// Include copies all errors and state of a binding result
func (br BindingResult) Include(other BindingResult) {
	for field, err := range other {
		br.Fail(field, err)
	}
}

// V is a helper type to quickly build variable maps for templates.
type V map[string]interface{}

// MarshalJSON implements the json.Marshaler interface.
func (v V) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}(v))
}

// JSON outputs the data encoded as JSON.
func JSON(w http.ResponseWriter, r *http.Request, code int, data interface{}) error {
	w.Header().Add("content-type", "application/json; charset=utf-8")
	w.WriteHeader(code)

	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")

	err := enc.Encode(data)
	return err
}

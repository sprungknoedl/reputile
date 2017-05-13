package middleware

import (
	"html/template"
	"net/http"

	"github.com/gorilla/context"
	"github.com/sprungknoedl/env"
	"github.com/sprungknoedl/reputile/handler"
	"github.com/sprungknoedl/reputile/lib"
)

func Templates(pattern string) func(http.Handler) http.Handler {
	debug := env.GetBool("debug")
	if debug {
		// parse template on each request
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tpl, err := template.ParseGlob(pattern)
				if err != nil {
					handler.HandleError(w, r, err)
					return
				}

				context.Set(r, lib.TemplateKey, tpl)
				next.ServeHTTP(w, r)
			})
		}
	} else {
		// cache parsed templates
		tpl := template.Must(template.ParseGlob(pattern))
		return WithValue(lib.TemplateKey, tpl)
	}
}

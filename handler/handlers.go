package handler

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/sprungknoedl/reputile/lib"
)

type V map[string]interface{}

func HTML(w http.ResponseWriter, r *http.Request, name string, data interface{}) {
	ctx := lib.NewContext(r)
	tpl := ctx.Value(lib.TemplateKey).(*template.Template)

	err := tpl.ExecuteTemplate(w, name, data)
	if err != nil {
		logrus.Printf("error during template %q: %v", name, err)
	}
}

func Text(w http.ResponseWriter, r *http.Request, data string) {
	w.Header().Add("content-type", "text/plain;charset=utf-8")
	w.Write([]byte(data))
}

func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "ERROR: %v", err)
}

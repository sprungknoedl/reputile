package handler

import "net/http"

func GetIndex(w http.ResponseWriter, r *http.Request) {
	HTML(w, r, "index.html", V{})
}

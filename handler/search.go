package handler

import (
	"net"
	"net/http"

	"github.com/sprungknoedl/reputile/lib"
	"github.com/sprungknoedl/reputile/model"
)

func GetSearch(w http.ResponseWriter, r *http.Request) {
	ctx := lib.NewContext(r)
	query := r.URL.Query().Get("q")
	result := []*model.Entry{}

	if query != "" {
		filter := map[string]string{}
		if ip := net.ParseIP(query); ip != nil {
			filter["ip4"] = ip.String()
		} else {
			filter["domain"] = query
		}

		entries := model.Find(ctx, filter)
		for entry := range entries {
			if entry.Err != nil {
				HandleError(w, r, entry.Err)
				return
			}

			result = append(result, entry)
		}
	}

	HTML(w, r, "search.html", V{
		"search":  query,
		"entries": result,
	})
}

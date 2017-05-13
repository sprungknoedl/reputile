package handler

import (
	"net/http"

	"github.com/sprungknoedl/reputile/lib"
	"github.com/sprungknoedl/reputile/lists"
	"github.com/sprungknoedl/reputile/model"
)

func GetLists(w http.ResponseWriter, r *http.Request) {
	ctx := lib.NewContext(r)
	cnt, err := model.CountEntries(ctx)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	HTML(w, r, "lists.html", V{
		"lists":   lists.Lists,
		"entries": cnt,
	})
}

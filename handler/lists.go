package handler

import (
	"net/http"
	"strconv"

	"github.com/sprungknoedl/reputile/cache"
	"github.com/sprungknoedl/reputile/lib"
	"github.com/sprungknoedl/reputile/lists"
	"github.com/sprungknoedl/reputile/model"
	"golang.org/x/net/context"
)

func GetLists(w http.ResponseWriter, r *http.Request) {
	ctx := lib.NewContext(r)
	downloads := cache.GetCounter(ctx, "stats:downloads")
	entries, err := cache.String(ctx, "stats:entries", func(ctx context.Context, key string) (string, error) {
		cnt, err := model.CountEntries(ctx)
		return strconv.Itoa(cnt), err
	})
	if err != nil {
		HandleError(w, r, err)
		return
	}

	HTML(w, r, "lists.html", V{
		"lists":     lists.Lists,
		"entries":   entries,
		"downloads": downloads,
	})
}

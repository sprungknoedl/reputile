package handler

import (
	"net/http"

	"github.com/sprungknoedl/reputile/cache"
	"github.com/sprungknoedl/reputile/lib"
)

func GetIndex(w http.ResponseWriter, r *http.Request) {
	ctx := lib.NewContext(r)
	size := cache.GetInt(ctx, "stats:size")

	HTML(w, r, "index.html", V{
		"size": size / 1024 / 1024, // in MiB
	})
}

package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/sprungknoedl/reputile/lib"
	"github.com/sprungknoedl/reputile/lists"
	"github.com/sprungknoedl/reputile/model"
	"golang.org/x/net/context"
)

func UpdateDatabase(w http.ResponseWriter, r *http.Request) {
	// check if update request is authorized
	token := viper.GetString("update_token")
	if r.Header.Get("authorization") != "Token "+token {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("update token required\n"))
		return
	}

	// create new background context & copy db handle
	db := lib.NewContext(r).Value(lib.DatabaseKey).(*model.Datastore)
	ctx := context.WithValue(context.Background(), lib.DatabaseKey, db)

	go func() {
		count := 0
		start := time.Now()

		// "convert" List to Iterator
		its := make([]lists.Iterator, len(lists.Lists))
		for i, list := range lists.Lists {
			its[i] = list
		}

		ch := lists.Combine(its...).Run(ctx)

		for entry := range ch {
			if entry.Err != nil {
				logrus.Errorf("(%s) failed to fetch entry: %v", entry.Source, entry.Err)
				return
			}

			count++
			err := model.Store(ctx, entry)
			if err != nil {
				logrus.Errorf("(%s) failed to store entry: %v", entry.Source, err)
				return
			}
		}

		model.Prune(ctx)
		logrus.Printf("added %d entries in %v", count, time.Since(start))
	}()

	fmt.Fprintf(w, "dispatched update job")
}
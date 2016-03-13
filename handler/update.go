package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/sprungknoedl/reputile/lib"
	"github.com/sprungknoedl/reputile/lists"
	"github.com/sprungknoedl/reputile/model"
	"golang.org/x/net/context"
)

func UpdateDatabase(w http.ResponseWriter, r *http.Request) {
	// create new background context & copy db handle
	db := lib.NewContext(r).Value(lib.DatabaseKey).(*model.Datastore)
	ctx := context.WithValue(context.Background(), lib.DatabaseKey, db)

	go func() {
		count := 0
		start := time.Now()
		ch := lists.Combine(lists.Lists...)(ctx)

		for entry := range ch {
			if entry.Err != nil {
				logrus.Errorf("failed to fetch entry: %v", entry.Err)
				return
			}

			count++
			err := model.Store(ctx, entry)
			if err != nil {
				logrus.Errorf("failed to store entry: %v", err)
				return
			}
		}

		logrus.Printf("added %d entries in %v", count, time.Since(start))
	}()

	fmt.Fprintf(w, "dispatched update job")
}

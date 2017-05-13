package main

import (
	"context"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/sprungknoedl/reputile/lib"
	"github.com/sprungknoedl/reputile/lists"
	"github.com/sprungknoedl/reputile/model"
)

func UpdateDatabase(db *model.Datastore) {
	// create new background context & copy db handle
	ctx := context.WithValue(context.Background(), lib.DatabaseKey, db)

	for {
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
		time.Sleep(1 * time.Hour)
	}
}

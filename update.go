package main

import (
	"context"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/sprungknoedl/reputile/lists"
)

func UpdateDatabase(db *Datastore) {
	// create new background context
	ctx := context.Background()

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
			err := db.Store(ctx, entry)
			if err != nil {
				logrus.Errorf("(%s) failed to store entry: %v", entry.Source, err)
				return
			}
		}

		db.Prune(ctx)
		logrus.Printf("added %d entries in %v", count, time.Since(start))
		time.Sleep(1 * time.Hour)
	}
}

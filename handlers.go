package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	gorilla "github.com/gorilla/context"
)

func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "ERROR: %v", err)
}

func GetDatabase(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(r)
	ch := Find(ctx, nil)

	w.Header().Add("content-type", "text/csv")
	writer := csv.NewWriter(w)

	for entry := range ch {
		if entry.Err != nil {
			HandleError(w, r, entry.Err)
			return
		}

		// csv format
		// source,domain,ip4,last,category,description
		writer.Write([]string{
			entry.Source,
			entry.Domain,
			entry.IP4,
			strconv.FormatInt(entry.Last.Unix(), 10),
			entry.Category,
			entry.Description,
		})
	}

	writer.Flush()
}

func UpdateDatabase(w http.ResponseWriter, r *http.Request) {
	// create new background context & copy db handle
	db := gorilla.Get(r, databaseKey).(*Datastore)
	ctx := context.WithValue(context.Background(), databaseKey, db)

	go func() {
		count := 0
		start := time.Now()
		ch := Combine(Lists...)(ctx)

		for entry := range ch {
			if entry.Err != nil {
				logrus.Errorf("failed to fetch entry: %v", entry.Err)
				return
			}

			count++
			err := Store(ctx, entry)
			if err != nil {
				logrus.Errorf("failed to store entry: %v", err)
				return
			}

		}

		logrus.Printf("added %d entries in %v", count, time.Since(start))
	}()

	fmt.Fprintf(w, "dispatched update job")
}

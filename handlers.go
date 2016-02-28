package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"net/url"
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

func FilterMap(values url.Values) map[string]string {
	m := make(map[string]string)
	for key, value := range values {
		if len(value) > 0 {
			m[key] = value[0]
		}
	}
	return m
}

func GetDatabase(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(r)

	filter := FilterMap(r.URL.Query())
	entries := Find(ctx, filter)

	w.Header().Add("content-type", "text/csv")
	writer := csv.NewWriter(w)

	for entry := range entries {
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

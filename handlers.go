package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"
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

func GetCron(w http.ResponseWriter, r *http.Request) {
	count := 0
	start := time.Now()

	ctx := NewContext(r)
	ch := Combine(Lists...)(ctx)

	for entry := range ch {
		if entry.Err != nil {
			HandleError(w, r, entry.Err)
			return
		}

		count++
		err := Store(ctx, entry)
		if err != nil {
			HandleError(w, r, err)
			return
		}
	}

	fmt.Fprintf(w, "added %d entries in %v", count, time.Since(start))
}

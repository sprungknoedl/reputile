package main

import (
	"fmt"
	"net/http"
	"time"
)

func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "ERROR: %v", err)
}

func GetDatabase(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(r)
	ch := Find(ctx, nil)

	for entry := range ch {
		if entry.Err != nil {
			HandleError(w, r, entry.Err)
			return
		}

		fmt.Fprintf(w, "%+v\n", entry)
	}
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

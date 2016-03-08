package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

func GetIndex(w http.ResponseWriter, r *http.Request) {
	HTML(w, r, "index.html", V{})
}

func GetLists(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(r)

	blacklists, err := Cache(ctx, "stats:blacklists", func(ctx context.Context, key string) (string, error) {
		cnt, err := CountSources(ctx)
		return strconv.Itoa(cnt), err
	})
	if err != nil {
		HandleError(w, r, err)
		return
	}

	entries, err := Cache(ctx, "stats:entries", func(ctx context.Context, key string) (string, error) {
		cnt, err := CountEntries(ctx)
		return strconv.Itoa(cnt), err
	})
	if err != nil {
		HandleError(w, r, err)
		return
	}

	downloads := GetCounter(ctx, "stats:downloads")

	HTML(w, r, "lists.html", V{
		"blacklists": blacklists,
		"entries":    entries,
		"downloads":  downloads,
	})
}

func CalculateDatabase(ctx context.Context, key string) (string, error) {
	start := time.Now()
	buffer := &bytes.Buffer{}

	r := ctx.Value(requestKey).(*http.Request)
	filter := FilterMap(r.URL.Query())
	entries := Find(ctx, filter)
	logrus.Printf("query %q; find took %v", key, time.Since(start))

	writer := csv.NewWriter(buffer)
	for entry := range entries {
		if entry.Err != nil {
			return "", entry.Err
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

	logrus.Printf("query %q; writing took %v", r.URL.RawQuery, time.Since(start))
	writer.Flush()

	err := writer.Error()
	return buffer.String(), err
}

func GetDatabase(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(r)
	list, err := Cache(ctx, r.URL.RawQuery, CalculateDatabase)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	IncrCounter(ctx, "stats:downloads")
	w.Header().Add("content-type", "text/plain;charset=utf-8")
	w.Write([]byte(list))
}

func UpdateDatabase(w http.ResponseWriter, r *http.Request) {
	// create new background context & copy db handle
	db := NewContext(r).Value(databaseKey).(*Datastore)
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

type V map[string]interface{}

func HTML(w http.ResponseWriter, r *http.Request, name string, data interface{}) {
	ctx := NewContext(r)
	tpl := ctx.Value(templateKey).(*template.Template)

	err := tpl.ExecuteTemplate(w, name, data)
	if err != nil {
		logrus.Printf("error during template %q: %v", name, err)
	}
}

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

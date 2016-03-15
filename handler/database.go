package handler

import (
	"bytes"
	"encoding/csv"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/sprungknoedl/reputile/cache"
	"github.com/sprungknoedl/reputile/lib"
	"github.com/sprungknoedl/reputile/model"
	"golang.org/x/net/context"
)

func GetDatabase(w http.ResponseWriter, r *http.Request) {
	ctx := lib.NewContext(r)
	key := "list:" + r.URL.RawQuery
	list, err := cache.String(ctx, key, CalculateDatabase)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	cache.Incr(ctx, "stats:downloads")
	Text(w, r, list)
}

func CalculateDatabase(ctx context.Context, key string) (string, error) {
	start := time.Now()
	buffer := &bytes.Buffer{}

	r := ctx.Value(lib.RequestKey).(*http.Request)
	filter := FilterMap(r.URL.Query())
	entries := model.Find(ctx, filter)
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
	cache.SetInt(ctx, "stats:size", buffer.Len())
	return buffer.String(), err
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

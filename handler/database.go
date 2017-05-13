package handler

import (
	"encoding/csv"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/sprungknoedl/reputile/lib"
	"github.com/sprungknoedl/reputile/model"
	"golang.org/x/net/context"
)

func GetDatabase(w http.ResponseWriter, r *http.Request) {
	ctx := lib.NewContext(r)

	w.Header().Add("content-type", "text/plain;charset=utf-8")
	err := CalculateDatabase(ctx, w, r.URL.Query())
	if err != nil {
		HandleError(w, r, err)
		return
	}
}

func CalculateDatabase(ctx context.Context, w io.Writer, query url.Values) error {
	start := time.Now()
	entries := model.Find(ctx, FilterMap(query))
	logrus.Printf("query %q; find took %v", query.Encode(), time.Since(start))

	writer := csv.NewWriter(w)
	for entry := range entries {
		if entry.Err != nil {
			return entry.Err
		}

		ip := ""
		if entry.IP != nil {
			ip = entry.IP.String()
		}

		// csv format
		// source,domain,ip4,last,category,description
		writer.Write([]string{
			entry.Source,
			entry.Domain,
			ip,
			strconv.FormatInt(entry.Last.Unix(), 10),
			entry.Category,
			entry.Description,
		})
	}

	logrus.Printf("query %q; writing took %v", query.Encode(), time.Since(start))
	writer.Flush()

	err := writer.Error()
	return err
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

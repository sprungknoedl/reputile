package main

import (
	"encoding/csv"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/nimbusec-oss/minion"
	"github.com/sprungknoedl/reputile/lists"
	"github.com/sprungknoedl/reputile/model"
)

func (app App) GetIndex(w http.ResponseWriter, r *http.Request) {
	app.HTML(w, r, http.StatusOK, "index.html", minion.V{})
}

func (app App) GetLists(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cnt, err := app.DB.CountEntries(ctx)
	if err != nil {
		app.Error(w, r, http.StatusInternalServerError, err)
		return
	}

	app.HTML(w, r, http.StatusOK, "lists.html", minion.V{
		"lists":   lists.Lists,
		"entries": cnt,
	})
}

func (app App) GetDatabase(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()

	start := time.Now()
	entries := app.DB.Find(ctx, FilterMap(query))
	logrus.Printf("query %q; find took %v", query.Encode(), time.Since(start))

	w.Header().Add("content-type", "text/plain;charset=utf-8")
	writer := csv.NewWriter(w)
	for entry := range entries {
		if entry.Err != nil {
			app.Error(w, r, http.StatusInternalServerError, entry.Err)
			return
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
	if err != nil {
		app.Error(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (app App) GetSearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query().Get("q")
	result := []*model.Entry{}

	if query != "" {
		filter := map[string]string{}

		if ip := net.ParseIP(query); ip != nil {
			filter["ip"] = ip.String()
		} else if _, ipnet, _ := net.ParseCIDR(query); ipnet != nil {
			filter["ip"] = ipnet.String()
		} else {
			filter["domain"] = query
		}

		logrus.Printf("filter = %+v", filter)
		entries := app.DB.Find(ctx, filter)
		for entry := range entries {
			if entry.Err != nil {
				app.Error(w, r, http.StatusInternalServerError, entry.Err)
				return
			}

			result = append(result, entry)
		}
	}

	app.HTML(w, r, http.StatusOK, "search.html", minion.V{
		"search":  query,
		"entries": result,
	})
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

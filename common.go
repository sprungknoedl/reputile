package main

import (
	"encoding/csv"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"

	"golang.org/x/net/context"
)

type List func(ctx context.Context) chan *Entry
type Translator func(row []string) *Entry

func ExtractHost(s string) string {
	if s == "" || s == "-" {
		return ""
	}

	if !(strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")) {
		s = "http://" + s
	}

	u, _ := url.Parse(s)
	return u.Host
}

func Combine(lists ...List) List {
	return func(ctx context.Context) chan *Entry {
		wg := sync.WaitGroup{}
		out := make(chan *Entry)

		// Start an output goroutine for each input channel in lists. output
		// copies values from c to out until c is closed, then calls wg.Done.
		output := func(c chan *Entry) {
			for entry := range c {
				out <- entry
			}
			wg.Done()
		}

		wg.Add(len(lists))
		for _, list := range lists {
			go output(list(ctx))
		}

		// Start a goroutine to close out once all the output goroutines are
		// done.  This must start after the wg.Add call.
		go func() {
			wg.Wait()
			close(out)
		}()

		return out
	}
}

func CSV(url string, fn Translator) List {
	return func(ctx context.Context) chan *Entry {
		out := make(chan *Entry)

		go func() {
			defer close(out)

			resp, err := http.Get(url)
			if err != nil {
				out <- SendError(err)
				return
			}

			defer resp.Body.Close()

			reader := csv.NewReader(resp.Body)
			reader.Comment = '#'

			for {
				row, err := reader.Read()
				if err == io.EOF {
					break
				}

				if err != nil {
					out <- SendError(err)
					break
				}

				select {
				case out <- fn(row):
				case <-ctx.Done():
					out <- SendError(ctx.Err())
					break
				}
			}

			logrus.Printf("update of %q finished", url)
		}()

		return out
	}
}

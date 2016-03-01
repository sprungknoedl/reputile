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

	u, err := url.Parse(s)
	if err != nil {
		//logrus.Printf("failed to parse %q: %v", s, err)
		return ""
	}

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

func TSV(url string, fn Translator) List {
	ctor := func(r io.Reader) *csv.Reader {
		reader := csv.NewReader(r)
		reader.Comma = '\t'
		reader.Comment = '#'
		reader.TrimLeadingSpace = true
		reader.FieldsPerRecord = -1
		return reader
	}

	return csvlist(url, fn, ctor)
}

func SSV(url string, fn Translator) List {
	ctor := func(r io.Reader) *csv.Reader {
		reader := csv.NewReader(r)
		reader.Comma = ' '
		reader.Comment = '#'
		reader.TrimLeadingSpace = true
		reader.FieldsPerRecord = -1
		return reader
	}

	return csvlist(url, fn, ctor)
}

func SSV2(url string, fn Translator) List {
	ctor := func(r io.Reader) *csv.Reader {
		reader := csv.NewReader(r)
		reader.Comma = ' '
		reader.Comment = '/'
		reader.TrimLeadingSpace = true
		reader.FieldsPerRecord = -1
		return reader
	}

	return csvlist(url, fn, ctor)
}

func CSV(url string, fn Translator) List {
	ctor := func(r io.Reader) *csv.Reader {
		reader := csv.NewReader(r)
		reader.Comment = '#'
		reader.TrimLeadingSpace = true
		reader.FieldsPerRecord = -1
		return reader
	}

	return csvlist(url, fn, ctor)
}

func CSV2(url string, fn Translator) List {
	ctor := func(r io.Reader) *csv.Reader {
		reader := csv.NewReader(r)
		reader.Comment = '/'
		reader.TrimLeadingSpace = true
		reader.FieldsPerRecord = -1
		return reader
	}

	return csvlist(url, fn, ctor)
}

func csvlist(url string, fn Translator, ctor func(io.Reader) *csv.Reader) List {
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
			reader := ctor(resp.Body)

			for {
				row, err := reader.Read()
				if err == io.EOF {
					break
				}

				if err != nil {
					out <- SendError(err)
					break
				}

				e := fn(row)
				if e == nil {
					continue
				}

				select {
				case out <- e:
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

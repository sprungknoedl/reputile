package lists

import (
	"encoding/csv"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/sprungknoedl/reputile/model"

	"golang.org/x/net/context"
)

var Lists []List

type List struct {
	Key         string
	Name        string
	URL         string
	Description string
	Generator   Generator
}

func Fetch(lists ...List) Generator {
	return func(ctx context.Context) chan *model.Entry {
		wg := sync.WaitGroup{}
		out := make(chan *model.Entry)

		// Start an output goroutine for each input channel in lists. output
		// copies values from c to out until c is closed, then calls wg.Done.
		output := func(list List, c chan *model.Entry) {
			for entry := range c {
				entry.Source = list.Key
				out <- entry
			}
			wg.Done()
		}

		wg.Add(len(lists))
		for _, list := range lists {
			go output(list, list.Generator(ctx))
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

type Generator func(ctx context.Context) chan *model.Entry

func Combine(generators ...Generator) Generator {
	return func(ctx context.Context) chan *model.Entry {
		wg := sync.WaitGroup{}
		out := make(chan *model.Entry)

		// Start an output goroutine for each input channel in generators. output
		// copies values from c to out until c is closed, then calls wg.Done.
		output := func(c chan *model.Entry) {
			for entry := range c {
				out <- entry
			}
			wg.Done()
		}

		wg.Add(len(generators))
		for _, gen := range generators {
			go output(gen(ctx))
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

type Translator func(row []string) *model.Entry

func CSV(url string, fn Translator) Generator {
	ctor := func(r io.Reader) *csv.Reader {
		reader := csv.NewReader(r)
		reader.Comment = '#'
		reader.TrimLeadingSpace = true
		reader.FieldsPerRecord = -1
		return reader
	}
	return GenericCSV(url, fn, ctor)
}

func SSV(url string, fn Translator) Generator {
	ctor := func(r io.Reader) *csv.Reader {
		reader := csv.NewReader(r)
		reader.Comma = ' '
		reader.Comment = '#'
		reader.TrimLeadingSpace = true
		reader.FieldsPerRecord = -1
		return reader
	}
	return GenericCSV(url, fn, ctor)
}

func TSV(url string, fn Translator) Generator {
	ctor := func(r io.Reader) *csv.Reader {
		reader := csv.NewReader(r)
		reader.Comma = '\t'
		reader.Comment = '#'
		reader.TrimLeadingSpace = true
		reader.FieldsPerRecord = -1
		return reader
	}
	return GenericCSV(url, fn, ctor)
}

func CStyleCSV(url string, fn Translator) Generator {
	ctor := func(r io.Reader) *csv.Reader {
		reader := csv.NewReader(r)
		reader.Comment = '/'
		reader.TrimLeadingSpace = true
		reader.FieldsPerRecord = -1
		return reader
	}
	return GenericCSV(url, fn, ctor)
}

func CStyleSSV(url string, fn Translator) Generator {
	ctor := func(r io.Reader) *csv.Reader {
		reader := csv.NewReader(r)
		reader.Comma = ' '
		reader.Comment = '/'
		reader.TrimLeadingSpace = true
		reader.FieldsPerRecord = -1
		return reader
	}
	return GenericCSV(url, fn, ctor)
}

func GenericCSV(url string, fn Translator, ctor func(io.Reader) *csv.Reader) Generator {
	return func(ctx context.Context) chan *model.Entry {
		out := make(chan *model.Entry)

		go func() {
			defer close(out)

			resp, err := http.Get(url)
			if err != nil {
				out <- model.SendError(err)
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
					out <- model.SendError(err)
					break
				}

				e := fn(row)
				if e == nil {
					continue
				}

				select {
				case out <- e:
				case <-ctx.Done():
					out <- model.SendError(ctx.Err())
					break
				}
			}

			logrus.Printf("update of %q finished", url)
		}()

		return out
	}
}

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

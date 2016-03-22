package lists

import (
	"encoding/csv"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/sprungknoedl/reputile/model"
	"golang.org/x/net/context"
)

var Lists []List

type Iterator interface {
	Run(context.Context) chan *model.Entry
}

type IteratorFunc func(ctx context.Context) chan *model.Entry

func (fn IteratorFunc) Run(ctx context.Context) chan *model.Entry {
	return fn(ctx)
}

type List struct {
	Key         string
	Name        string
	URL         string
	Description string
	Iterator    Iterator
}

// Run runs the iterator defined for this List. It also sets
// list shared attributes to each entry
func (l List) Run(ctx context.Context) chan *model.Entry {
	c := l.Iterator.Run(ctx)
	out := make(chan *model.Entry)

	go func() {
		defer close(out)
		for entry := range c {
			entry.Source = l.Key
			if entry.Domain == "" && entry.IP == nil {
				logrus.Warnf("skipping invalid entry: %+v", entry)
				continue
			}

			out <- entry
		}
	}()

	return out
}

func Combine(its ...Iterator) Iterator {
	return IteratorFunc(func(ctx context.Context) chan *model.Entry {
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

		wg.Add(len(its))
		for _, it := range its {
			go output(it.Run(ctx))
		}

		// Start a goroutine to close out once all the output goroutines are
		// done.  This must start after the wg.Add call.
		go func() {
			wg.Wait()
			close(out)
		}()

		return out
	})
}

type Translator func(row []string) *model.Entry

func CSV(url string, fn Translator) Iterator {
	ctor := func(r io.Reader) *csv.Reader {
		reader := csv.NewReader(r)
		reader.Comment = '#'
		reader.TrimLeadingSpace = true
		reader.FieldsPerRecord = -1
		return reader
	}
	return GenericCSV(url, fn, ctor)
}

func SSV(url string, fn Translator) Iterator {
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

func TSV(url string, fn Translator) Iterator {
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

func CStyleCSV(url string, fn Translator) Iterator {
	ctor := func(r io.Reader) *csv.Reader {
		reader := csv.NewReader(r)
		reader.Comment = '/'
		reader.TrimLeadingSpace = true
		reader.FieldsPerRecord = -1
		return reader
	}
	return GenericCSV(url, fn, ctor)
}

func CStyleSSV(url string, fn Translator) Iterator {
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

func GenericCSV(url string, fn Translator, ctor func(io.Reader) *csv.Reader) Iterator {
	return IteratorFunc(func(ctx context.Context) chan *model.Entry {
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
	})
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
		logrus.Warnf("failed to parse %s: %v", s, err)
		return ""
	}

	host := u.Host
	if strings.ContainsRune(host, ':') {
		host, _, err = net.SplitHostPort(host)
		if err != nil {
			logrus.Warnf("failed to parse %s: %v", u.Host, err)
			return ""
		}
	}

	return host
}

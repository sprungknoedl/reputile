package main

import (
	"html/template"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/carbocation/interpose"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

var Lists []List

func config() {
	viper.AutomaticEnv()
	viper.SetDefault("debug", false)
	viper.SetDefault("port", "3000")
	viper.SetDefault("database_url", "postgres://localhost/reputile")
}

func main() {
	config()
	databaseURL := viper.GetString("database_url")
	redisURL := viper.GetString("redis_url")
	port := viper.GetString("port")

	store, err := NewDatastore(databaseURL)
	if err != nil {
		logrus.Fatal(err)
	}

	cache, err := redis.DialURL(redisURL)
	if err != nil {
		logrus.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/", GetIndex)
	router.HandleFunc("/lists", GetLists)
	router.HandleFunc("/lists/database.txt", GetDatabase)
	router.HandleFunc("/_internal/update", UpdateDatabase)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("static")))

	middle := interpose.New()
	middle.Use(Templates("templates/*.html"))
	middle.Use(WithValue(cacheKey, cache))
	middle.Use(WithValue(databaseKey, store))
	middle.UseHandler(router)

	logrus.Printf("listening on :%s", port)
	err = http.ListenAndServe(":"+port, middle)
	if err != nil {
		logrus.Fatal(err)
	}
}

func Templates(pattern string) func(http.Handler) http.Handler {
	debug := viper.GetBool("debug")
	if debug {
		// parse template on each request
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tpl, err := template.ParseGlob(pattern)
				if err != nil {
					HandleError(w, r, err)
					return
				}

				context.Set(r, templateKey, tpl)
				next.ServeHTTP(w, r)
			})
		}

	} else {
		// cache parsed templates
		tpl := template.Must(template.ParseGlob(pattern))
		return WithValue(templateKey, tpl)
	}
}

func WithValue(key, val interface{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			context.Set(r, key, val)
			next.ServeHTTP(w, r)
		})
	}
}

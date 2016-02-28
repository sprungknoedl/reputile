package main

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/carbocation/interpose"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

var Lists []List

func config() {
	viper.AutomaticEnv()
	viper.SetDefault("port", "3000")
	viper.SetDefault("database_url", "postgres://localhost/reputile")
}

func main() {
	config()
	db := viper.GetString("database_url")
	port := viper.GetString("port")

	conn, err := NewDatastore(db)
	if err != nil {
		logrus.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/database.csv", GetDatabase)
	router.HandleFunc("/_internal/update", UpdateDatabase)

	middle := interpose.New()
	middle.Use(WithValue(databaseKey, conn))
	middle.UseHandler(router)

	logrus.Printf("listening on :%s", port)
	err = http.ListenAndServe(":"+port, middle)
	if err != nil {
		logrus.Fatal(err)
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

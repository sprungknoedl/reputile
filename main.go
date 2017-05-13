package main

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/nimbusec-oss/minion"
	"github.com/sprungknoedl/env"
)

type App struct {
	*minion.Minion
	DB *Datastore
}

func main() {
	port := env.GetString("port")
	dburl := env.GetString("db")
	db, err := NewDatastore(dburl)
	if err != nil {
		logrus.Fatal(err)
	}

	app := App{
		Minion: minion.NewMinion(),
		DB:     db,
	}

	router := mux.NewRouter()
	router.HandleFunc("/", app.GetIndex)
	router.HandleFunc("/lists", app.GetLists)
	router.HandleFunc("/lists/database.txt", app.GetDatabase)
	router.HandleFunc("/search", app.GetSearch)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("static")))

	if !env.GetBool("disable.update") {
		go UpdateDatabase(db)
	}

	logrus.Printf("listening on :%s", port)
	err = http.ListenAndServe(":"+port, router)
	if err != nil {
		logrus.Fatal(err)
	}
}

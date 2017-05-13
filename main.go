package main

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/carbocation/interpose"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/sprungknoedl/env"
	"github.com/sprungknoedl/reputile/handler"
	"github.com/sprungknoedl/reputile/lib"
	"github.com/sprungknoedl/reputile/middleware"
	"github.com/sprungknoedl/reputile/model"
)

func main() {
	port := env.GetString("port")
	databaseURL := env.GetString("db")
	store, err := model.NewDatastore(databaseURL)
	if err != nil {
		logrus.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/", handler.GetIndex)
	router.HandleFunc("/lists", handler.GetLists)
	router.HandleFunc("/lists/database.txt", handler.GetDatabase)
	router.HandleFunc("/search", handler.GetSearch)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("static")))

	middle := interpose.New()
	middle.Use(middleware.Templates("templates/*.html"))
	middle.Use(middleware.WithValue(lib.DatabaseKey, store))
	middle.UseHandler(router)

	go UpdateDatabase(store)

	logrus.Printf("listening on :%s", port)
	err = http.ListenAndServe(":"+port, middle)
	if err != nil {
		logrus.Fatal(err)
	}
}

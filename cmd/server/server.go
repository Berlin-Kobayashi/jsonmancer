package main

import (
	"net/http"
	"github.com/DanShu93/jsonmancer/mongo"
	"github.com/DanShu93/jsonmancer/storage"
	"gopkg.in/mgo.v2"
	"github.com/DanShu93/jsonmancer/uuid"
)

func main() {
	mongoURL := "db:27017"
	mongoDB := "jsonmancer"

	storageService, err := build(mongoURL, mongoDB)
	if err != nil {
		panic(err)
	}

	http.Handle("/", storageService)

	http.ListenAndServe(":80", nil)
}

func build(mongoURL, mongoDB string) (http.Handler, error) {
	session, err := mgo.Dial(mongoURL)
	if err != nil {
		panic(err)
	}

	err = session.DB(mongoDB).DropDatabase()
	if err != nil {
		panic(err)
	}

	repository, err := mongo.New(
		mongoURL,
		mongoDB,
	)
	if err != nil {
		panic(err)
	}

	store, err := storage.New(storage.FixtureEntities, repository, uuid.V4{})
	if err != nil {
		panic(err)
	}

	return storage.Service{Storage: store}, nil
}

package main

import (
	"net/http"
	"github.com/DanShu93/jsonmancer/mongo"
)

type StorageService struct {
}

func (s StorageService) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

}

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
	_, err := mongo.New(
		mongoURL,
		mongoDB,
	)
	if err != nil {
		return nil, err
	}

	// TODO use fixture service
	return StorageService{}, nil
}

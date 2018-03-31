package storage

import (
	"regexp"
	"net/http"
	"fmt"
	"encoding/json"
	"io/ioutil"
)

const ActionExpand = "expand"
const ActionReferencedBy = "referenced-by"
const Meta = "meta"
const MetaActionSwaggerFile = "swagger"

var entityNameRegex = regexp.MustCompile("^/([^/]+)/.*$")
var indexRegex = regexp.MustCompile("^.*/([^/]+)$")
var actionRegex = regexp.MustCompile("^/[^/]+/([^/]+)$")
var indexedActionRegex = regexp.MustCompile("^/[^/]+/([^/]+)/[^/]+$")
var indexedEntityNameRegex = regexp.MustCompile("^/([^/]+)$")

var pathRegex = regexp.MustCompilePOSIX("^/[^/]+/?[^/]*/?[^/]*$")

type Service struct {
	Storage Storage
	Info    Info
}

type Info struct {
	Title, Version string
}

func (s Service) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")

	entityName := s.getEntityName(r)

	if !pathRegex.Match([]byte(r.URL.Path)) {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	index := s.getIndex(r)
	action := s.getAction(r)

	if entityName == Meta {
		switch action {
		case MetaActionSwaggerFile:
			s.GetSwaggerFile(rw, r)
		}
	} else {
		switch r.Method {
		case http.MethodGet:
			switch action {
			case ActionExpand:
				s.expand(rw, r, entityName, index)
			case ActionReferencedBy:
				s.getReferencedBy(rw, r, entityName, index)
			default:
				if index == "" || index == entityName {
					s.getAll(rw, r, entityName)
				} else {
					s.get(rw, r, entityName, index)
				}
			}
		case http.MethodPost:
			s.post(rw, r, entityName)
		case http.MethodPut:
			s.put(rw, r, entityName, index)
		case http.MethodDelete:
			s.delete(rw, r, entityName, index)
		default:
			rw.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func (s Service) GetSwaggerFile(rw http.ResponseWriter, r *http.Request) {
	response, err := CreateSwaggerFile(s.Storage.entities, s.Info, r.Host)
	if err != nil {
		fmt.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Write([]byte(response))
}

func (s Service) get(rw http.ResponseWriter, r *http.Request, entityName string, index string) {
	resource, err := s.Storage.Read(entityName, index)
	if err != nil {
		fmt.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(resource)
	if err != nil {
		fmt.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Write(response)
}

func (s Service) getAll(rw http.ResponseWriter, r *http.Request, entityName string) {
	resource, err := s.Storage.ReadAll(entityName, Query{})
	if err != nil {
		fmt.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(resource)
	if err != nil {
		fmt.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Write(response)
}

func (s Service) expand(rw http.ResponseWriter, r *http.Request, entityName string, index string) {
	resource, err := s.Storage.ReadAndExpand(entityName, index)
	if err != nil {
		fmt.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(resource)
	if err != nil {
		fmt.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Write(response)
}

func (s Service) getReferencedBy(rw http.ResponseWriter, r *http.Request, entityName string, index string) {
	resource, err := s.Storage.GetReferencedBy(entityName, index)
	if err != nil {
		fmt.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(resource)
	if err != nil {
		fmt.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Write(response)
}

func (s Service) post(rw http.ResponseWriter, r *http.Request, entityName string) {
	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	resource, err := s.Storage.CreateFromJSON(entityName, string(content))
	if err != nil {
		fmt.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(resource)
	if err != nil {
		fmt.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Write(response)
}

func (s Service) put(rw http.ResponseWriter, r *http.Request, entityName string, index string) {
	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	resource, err := s.Storage.UpdateFromJSON(entityName, string(content))
	if err != nil {
		fmt.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(resource)
	if err != nil {
		fmt.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Write(response)
}

func (s Service) delete(rw http.ResponseWriter, r *http.Request, entityName string, index string) {
	err := s.Storage.Purge(entityName, index)
	if err != nil {
		fmt.Println(err)
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}

func (s Service) getAction(r *http.Request) string {
	regex := actionRegex

	if !regex.Match([]byte(r.URL.Path)) {
		regex = indexedActionRegex
	}
	if !regex.Match([]byte(r.URL.Path)) {
		return ""
	}

	return string(regex.ReplaceAll([]byte(r.URL.Path), []byte("$1")))
}

func (s Service) getIndex(r *http.Request) string {
	if !indexRegex.Match([]byte(r.URL.Path)) {
		return ""
	}

	return string(indexRegex.ReplaceAll([]byte(r.URL.Path), []byte("$1")))
}

func (s Service) getEntityName(r *http.Request) string {
	regex := entityNameRegex

	if !regex.Match([]byte(r.URL.Path)) {
		regex = indexedEntityNameRegex
	}

	return string(regex.ReplaceAll([]byte(r.URL.Path), []byte("$1")))
}

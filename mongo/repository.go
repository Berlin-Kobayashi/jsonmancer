package mongo

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/DanShu93/jsonmancer/storage"
)

type Repository struct {
	database *mgo.Database
}

func New(url, db string) (Repository, error) {
	session, err := mgo.Dial(url)

	if err != nil {
		return Repository{}, storage.DBError{Message: err.Error()}
	}

	database := session.DB(db)

	return Repository{database: database}, nil
}

func (s Repository) Create(collectionName string, data interface{}) error {
	err := s.database.C(collectionName).Insert(data)
	if err != nil {
		return storage.DBError{Message: err.Error()}
	}

	return nil
}

func (s Repository) Read(collectionName, id string, result interface{}) error {
	q := s.database.C(collectionName).Find(bson.M{"_id": id})

	n, err := q.Count()
	if err != nil {
		return storage.DBError{Message: err.Error()}
	}

	if n == 0 {
		return storage.NotFound{Entity: collectionName, ID: id}
	}

	err = q.One(result)
	if err != nil {
		return storage.DBError{Message: err.Error()}
	}

	return nil
}

func (s Repository) Update(collectionName, id string, data interface{}) error {
	err := s.database.C(collectionName).Update(bson.M{"_id": id}, data)
	if err != nil {
		return storage.DBError{Message: err.Error()}
	}

	return nil
}

func (s Repository) Delete(collectionName, id string) error {
	err := s.database.C(collectionName).Remove(bson.M{"_id": id})
	if err != nil {
		return storage.DBError{Message: err.Error()}
	}

	return nil
}

func (s Repository) ReadAll(collectionName string, query storage.Query, result interface{}) error {
	mq := createMongoQuery(query)
	q := s.database.C(collectionName).Find(mq)

	n, err := q.Count()
	if err != nil {
		return storage.DBError{Message: err.Error()}
	}

	if n == 0 {
		return storage.NoMatch{Entity:collectionName,Query:query}
	}

	err = q.All(result)
	if err != nil {
		return storage.DBError{Message: err.Error()}
	}

	return nil
}

func createMongoQuery(q storage.Query) bson.M {
	mq := bson.M{}
	or := make([]bson.M, 0)
	and := make([]bson.M, 0)
	for k, v := range q.Q {
		if k == "ID" {
			k = "_id"
		}

		fieldQueries := make([]bson.M, len(v.Values))
		for i, currentValue := range v.Values {
			fieldQueries[i] = bson.M{k: currentValue}
		}

		switch v.Kind {
		case storage.QueryAnd:
			and = append(and, fieldQueries...)
		case storage.QueryOr:
			or = append(or, fieldQueries...)
		case storage.QueryContains:
			mq[k] = bson.M{"$in": v.Values}
		}
	}

	if len(or) != 0 {
		mq["$or"] = or
	}
	if len(and) != 0 {
		mq["$and"] = and
	}

	return mq
}

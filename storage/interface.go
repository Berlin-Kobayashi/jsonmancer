package storage

type Repository interface {
	Create(collectionName string, data interface{}) error
	Read(collectionName string, id string, result interface{}) error
	Update(collectionName string, id string, data interface{}) error
	Delete(collectionName string, id string) error
	ReadAll(collectionName string, query Query, result interface{}) error
}

type IDGenerator interface {
	Generate() string
}

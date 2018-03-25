package storage

const (
	QueryAnd      kind = iota
	QueryOr
	QueryContains
)

type Query struct {
	Q map[string]FieldQuery
}

type FieldQuery struct {
	Kind   kind
	Values []interface{}
}

type kind uint

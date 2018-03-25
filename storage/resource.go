package storage

import "reflect"

type Entity struct {
	Name       string
	Data       reflect.Type
	References map[string]Entity
}

func (e Entity) New() Resource {
	references := make(map[string][]Resource, len(e.References))
	for k := range e.References {
		if v, ok := references[k]; ok {
			references[k] = v
		} else {
			references[k] = []Resource{}
		}
	}
	data := reflect.New(e.Data).Interface()

	return Resource{Data: data, References: references}
}

type Resource struct {
	ID         string
	Data       interface{}
	References map[string][]Resource
}

type CollapsedResource struct {
	ID         string `bson:"_id"`
	Data       interface{}
	References map[string][]string
}

func (r Resource) Collapse() CollapsedResource {
	result := CollapsedResource{}
	result.ID = r.ID
	result.Data = r.Data

	references := make(map[string][]string, len(r.References))

	for k, v := range r.References {
		ids := make([]string, len(v))
		for i, r := range v {
			ids[i] = r.ID
		}

		references[k] = ids
	}

	result.References = references

	return result
}

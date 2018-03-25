package storage

import (
	"reflect"
)

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

	return Resource{Data: data, References: references, entity: e}
}

func (e Entity) Read(repository Repository, id string) (CollapsedResource, error) {
	result := e.New().Collapse()

	err := repository.Read(e.Name, id, &result)
	if err != nil {
		return CollapsedResource{}, err
	}

	result.entity = e

	return result, nil
}

type Resource struct {
	ID         string
	Data       interface{}
	References map[string][]Resource
	entity     Entity
}

type CollapsedResource struct {
	ID         string `bson:"_id"`
	Data       interface{}
	References map[string][]string
	entity     Entity
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
	result.entity = r.entity

	return result
}

func (r *CollapsedResource) Expand(repository Repository) (Resource, error) {
	resource := Resource{}
	resource.ID = r.ID
	resource.Data = r.Data
	resource.entity = r.entity
	resource.References = make(map[string][]Resource, len(r.References))

	for relationName, references := range r.References {
		referenceEntity := r.entity.References[relationName]

		resource.References[relationName] = make([]Resource, len(references))
		for i, reference := range references {
			result, err := referenceEntity.Read(repository, reference)
			if err != nil {
				return Resource{}, err
			}

			referencedResource, err := result.Expand(repository)
			if err != nil {
				return Resource{}, err
			}

			resource.References[relationName][i] = referencedResource
		}
	}

	return resource, nil
}

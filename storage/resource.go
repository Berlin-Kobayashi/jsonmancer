package storage

import (
	"reflect"
	"fmt"
)

type Entities struct {
	entitiesByName map[string]Entity
	referencedBy   map[string]map[string][]string
}

func NewEntities(entities []Entity) (Entities, error) {
	entityMap, err := mapEntities(entities)
	if err != nil {
		return Entities{}, err
	}

	referencedBy, err := getReferencedBy(entityMap)
	if err != nil {
		return Entities{}, err
	}

	return Entities{
		entitiesByName: entityMap,
		referencedBy:   referencedBy,
	}, nil
}

func mapEntities(entityDefinition []Entity) (map[string]Entity, error) {
	entityMap := make(map[string]Entity, len(entityDefinition))
	for _, v := range entityDefinition {
		if _, ok := entityMap[v.Name]; ok {
			return nil, fmt.Errorf("entitiy name %q i not unique", v.Name)
		}
		entityMap[v.Name] = v
	}

	return entityMap, nil
}

func getReferencedBy(entities map[string]Entity) (map[string]map[string][]string, error) {
	referenceBy := make(map[string]map[string][]string, len(entities))
	for name := range entities {
		referenceBy[name] = map[string][]string{}
	}

	for entityName, entity := range entities {
		for relationName, reference := range entity.References {
			if _, ok := entities[reference.Name]; !ok {
				return nil, fmt.Errorf("entitiy %q is referenced but unknown", reference.Name)
			}

			if _, ok := referenceBy[entityName][reference.Name]; !ok {
				referenceBy[reference.Name][entityName] = []string{}
			}

			referenceBy[reference.Name][entityName] = append(referenceBy[reference.Name][entityName], relationName)
		}
	}

	return referenceBy, nil
}

func (e *Entities) CreateReferencedByMap(entityName string) (map[string]map[string][]string, error) {
	references, ok := e.referencedBy[entityName]
	if !ok {
		return nil, UndefinedEntity{entityName}
	}

	result := make(map[string]map[string][]string, len(references))

	for k, v := range references {
		result[k] = make(map[string][]string, len(v))
		for _, relationName := range v {
			result[k][relationName] = []string{}
		}
	}

	return result, nil
}

type Entity struct {
	Name       string
	Data       reflect.Type
	References map[string]Entity
}

func (e Entity) New() Resource {
	data := reflect.New(e.Data).Interface()

	references := make(map[string][]Resource, len(e.References))
	for k := range e.References {
		references[k] = []Resource{}
	}

	return Resource{Data: data, References: references, entity: e}
}

type Resource struct {
	ID         string                `json:"id"`
	Data       interface{}           `json:"data"`
	References map[string][]Resource `json:"references"`
	entity     Entity
}

type CollapsedResource struct {
	ID         string              `bson:"_id" json:"id"`
	Data       interface{}         `json:"data"`
	References map[string][]string `json:"references"`
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

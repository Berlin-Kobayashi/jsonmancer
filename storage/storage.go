package storage

import (
	"fmt"
	"encoding/json"
)

type Storage struct {
	entities     map[string]Entity
	referencedBy map[string][]string
	repository   Repository
	idGenerator  IDGenerator
}

func New(entities []Entity, repository Repository, idGenerator IDGenerator) (Storage, error) {
	entityMap, err := mapEntities(entities)
	if err != nil {
		return Storage{}, err
	}

	referencedBy, err := getReferencedBy(entityMap)
	if err != nil {
		return Storage{}, err
	}

	return Storage{
		entities:     entityMap,
		repository:   repository,
		idGenerator:  idGenerator,
		referencedBy: referencedBy,
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

func getReferencedBy(entities map[string]Entity) (map[string][]string, error) {
	referenceBy := make(map[string][]string, len(entities))
	for name := range entities {
		referenceBy[name] = []string{}
	}

	for entityName, entity := range entities {
		for relationName, reference := range entity.References {
			if _, ok := entities[reference.Name]; !ok {
				return nil, fmt.Errorf("entitiy %q is referenced but unknown", reference.Name)
			}

			referenceBy[entityName] = append(referenceBy[entityName], relationName)
		}
	}

	return referenceBy, nil
}

func (s *Storage) Create(entityName, jsonDocument string) (CollapsedResource, error) {
	resource, err := s.createCollapsedResource(entityName, jsonDocument)
	if err != nil {
		return CollapsedResource{}, err
	}

	_, err = resource.Expand(s.repository)
	if err != nil {
		return CollapsedResource{}, err
	}

	resource.ID = s.idGenerator.Generate()

	err = s.repository.Create(entityName, resource)
	if err != nil {
		return CollapsedResource{}, err
	}

	return resource, nil
}

func (s *Storage) createCollapsedResource(entityName, jsonDocument string) (CollapsedResource, error) {
	entity, ok := s.entities[entityName]
	if !ok {
		return CollapsedResource{}, UndefinedEntity{entityName}
	}

	resource := entity.New().Collapse()

	err := json.Unmarshal([]byte(jsonDocument), &resource)
	if err != nil {
		return CollapsedResource{}, err
	}

	return resource, nil
}

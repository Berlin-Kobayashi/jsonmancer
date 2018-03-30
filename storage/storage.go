package storage

import (
	"fmt"
	"encoding/json"
)

type Storage struct {
	entities     map[string]Entity
	referencedBy map[string]map[string][]string
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

func (s *Storage) Create(entityName, jsonDocument string) (CollapsedResource, error) {
	resource, err := s.createCollapsedResource(entityName, jsonDocument)
	if err != nil {
		return CollapsedResource{}, err
	}

	_, err = s.Expand(resource)
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

func (s *Storage) ReadAndExpand(entityName, id string) (Resource, error) {
	collapsedResource, err := s.Read(entityName, id)
	if err != nil {
		return Resource{}, err
	}

	return s.Expand(collapsedResource)
}

func (s *Storage) Read(entityName, id string) (CollapsedResource, error) {
	entity, ok := s.entities[entityName]
	if !ok {
		return CollapsedResource{}, UndefinedEntity{entityName}
	}

	result := entity.New().Collapse()

	err := s.repository.Read(entity.Name, id, &result)
	if err != nil {
		return CollapsedResource{}, err
	}

	result.entity = entity

	return result, nil
}

func (s *Storage) Expand(collapsedResource CollapsedResource) (Resource, error) {
	resource := Resource{}
	resource.ID = collapsedResource.ID
	resource.Data = collapsedResource.Data
	resource.entity = collapsedResource.entity
	resource.References = make(map[string][]Resource, len(collapsedResource.References))

	for relationName, references := range collapsedResource.References {
		referenceEntity := collapsedResource.entity.References[relationName]

		resource.References[relationName] = make([]Resource, len(references))
		for i, reference := range references {
			result, err := s.Read(referenceEntity.Name, reference)
			if err != nil {
				return Resource{}, err
			}

			referencedResource, err := s.Expand(result)
			if err != nil {
				return Resource{}, err
			}

			resource.References[relationName][i] = referencedResource
		}
	}

	return resource, nil
}

func (s *Storage) GetReferencedBy(entityName, id string) (map[string]map[string][]string, error) {
	referencedBy, err := s.createReferencedByMap(entityName)
	if err != nil {
		return nil, err
	}

	for referencingEntityName, references := range referencedBy {
		for relationName := range references {
			query := Query{Q: make(map[string]FieldQuery, len(references))}
			query.Q["references."+relationName] = FieldQuery{Kind: QueryContains, Values: []interface{}{id}}
			result := []CollapsedResource{}
			err = s.repository.ReadAll(referencingEntityName, query, &result)
			if err != nil {
				return nil, err
			}

			for _, row := range result {
				referencedBy[referencingEntityName][relationName] = append(referencedBy[referencingEntityName][relationName], row.ID)
			}
		}
	}

	return referencedBy, nil
}

func (s *Storage) createReferencedByMap(entityName string) (map[string]map[string][]string, error) {
	references, ok := s.referencedBy[entityName]
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

package storage

import (
	"encoding/json"
)

type Storage struct {
	entities     Entities
	repository   Repository
	idGenerator  IDGenerator
}

func New(entities []Entity, repository Repository, idGenerator IDGenerator) (Storage, error) {
	validatedEntities, err := NewEntities(entities)
	if err != nil {
		return Storage{}, err
	}

	return Storage{
		entities:     validatedEntities,
		repository:   repository,
		idGenerator:  idGenerator,
	}, nil
}

func (s *Storage) CreateFromJSON(entityName, jsonDocument string) (CollapsedResource, error) {
	resource, err := s.createCollapsedResourceFromJSON(entityName, jsonDocument)
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

func (s *Storage) UpdateFromJSON(entityName, jsonDocument string) (CollapsedResource, error) {
	resource, err := s.createCollapsedResourceFromJSON(entityName, jsonDocument)
	if err != nil {
		return CollapsedResource{}, err
	}

	_, err = s.Expand(resource)
	if err != nil {
		return CollapsedResource{}, err
	}

	err = s.Update(resource)
	if err != nil {
		return CollapsedResource{}, err
	}

	return resource, nil
}

func (s *Storage) Update(collapsedResource CollapsedResource) error {
	return s.repository.Update(collapsedResource.entity.Name, collapsedResource.ID, collapsedResource)
}

func (s *Storage) ReadAndExpand(entityName, id string) (Resource, error) {
	collapsedResource, err := s.Read(entityName, id)
	if err != nil {
		return Resource{}, err
	}

	return s.Expand(collapsedResource)
}

func (s *Storage) Read(entityName, id string) (CollapsedResource, error) {
	entity, ok := s.entities.entitiesByName[entityName]
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

func (s *Storage) ReadAll(entityName string, query Query) ([]CollapsedResource, error) {
	entity, ok := s.entities.entitiesByName[entityName]
	if !ok {
		return nil, UndefinedEntity{entityName}
	}

	result := []CollapsedResource{}
	err := s.repository.ReadAll(entity.Name, query, &result)
	if err != nil {
		return nil, err
	}

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
	referencedBy, err := s.entities.CreateReferencedByMap(entityName)
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

// Deletes the resource and all references to it.
func (s *Storage) Purge(entityName, id string) error {
	referencedBy, err := s.GetReferencedBy(entityName, id)
	if err != nil {
		return err
	}

	for referencingEntityName, references := range referencedBy {
		for relationName, referenceIDs := range references {
			for _, referenceID := range referenceIDs {
				reference, err := s.Read(referencingEntityName, referenceID)
				if err != nil {
					return err
				}

				newReferences := []string{}
				for _, v := range reference.References[relationName] {
					if v != id {
						newReferences = append(newReferences, v)
					}
				}

				reference.References[relationName] = newReferences
				err = s.Update(reference)
				if err != nil {
					return err
				}
			}
		}
	}

	return s.repository.Delete(entityName, id)
}

func (s *Storage) Delete(entityName, id string) error {
	return s.repository.Delete(entityName, id)
}

func (s *Storage) createCollapsedResourceFromJSON(entityName, jsonDocument string) (CollapsedResource, error) {
	entity, ok := s.entities.entitiesByName[entityName]
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

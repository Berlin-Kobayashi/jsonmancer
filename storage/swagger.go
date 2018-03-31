package storage

import (
	"encoding/json"
)

func CreateSwaggerFile(entities Entities, info Info, host string) (string,error) {
	paths := map[string]interface{}{
		"/meta/swagger": map[string]interface{}{
			"get": map[string]interface{}{
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "This swagger file",
					},
				},
			},
		},
	}

	definitions := map[string]interface{}{}

	for entityName, entity := range entities.entitiesByName {
		paths["/"+entityName] = map[string]interface{}{
			"get": map[string]interface{}{
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"schema": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"$ref": "#/definitions/" + entityName,
							},
						},
					},
				},
			},
		}

		definitions[entityName] = entity.New().Collapse()
	}

	swagger := map[string]interface{}{
		"swagger": "2.0",
		"info": map[string]interface{}{
			"version": info.Version,
			"title":   info.Title,
		},
		"host":  host,
		"paths": paths,
		"definitions" : definitions,
	}

	content, err := json.Marshal(swagger)
	if err != nil {
		return "",err
	}

	return string(content), nil
}

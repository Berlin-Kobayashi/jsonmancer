package storage

import (
	"encoding/json"
	"fmt"
)

func CreateSwaggerFile(entities Entities, info Info, host string) (string, error) {
	paths := map[string]interface{}{
		fmt.Sprintf("/%s/%s", Meta, MetaActionSwaggerFile): map[string]interface{}{
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
						"description": "All matching " + entityName,
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

		schemaReference := map[string]interface{}{
			"$ref": "#/definitions/" + entityName,
		}

		pathParameterName := entityName + "Id"
		pathParameter := map[string]interface{}{
			"name":        pathParameterName,
			"in":          "path",
			"description": "ID of the " + entityName,
			"required":    true,
			"type":        "string",
		}

		bodyParameter := map[string]interface{}{
			"name":        "body",
			"in":          "body",
			"description": "ID of the " + entityName,
			"required":    true,
			"schema":      schemaReference,
		}

		paths["/"+entityName+"/{"+pathParameterName+"}"] = map[string]interface{}{
			"parameters": []interface{}{
				pathParameter,
			},
			"get": map[string]interface{}{
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "A single " + entityName,
						"schema":      schemaReference,
					},
				},
			},
			"post": map[string]interface{}{
				"parameters": []interface{}{
					bodyParameter,
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "The created " + entityName,
						"schema":      schemaReference,
					},
				},
			},
			"put": map[string]interface{}{
				"parameters": []interface{}{
					bodyParameter,
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "The updated " + entityName,
						"schema":      schemaReference,
					},
				},
			},
			"delete": map[string]interface{}{
				"responses": map[string]interface{}{
					"204": map[string]interface{}{
						"description": "No content",
					},
				},
			},
		}

		expandedEntityName := entityName + "Expanded"

		paths[fmt.Sprintf("/%s/%s/{%s}", entityName, ActionExpand, pathParameterName)] = map[string]interface{}{
			"get": map[string]interface{}{
				"parameters": []interface{}{
					pathParameter,
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "The expanded " + entityName,
						"schema": map[string]interface{}{
							"$ref": "#/definitions/" + expandedEntityName,
						},
					},
				},
			},
		}

		referencedByEntityName := entityName + "ReferencedBy"

		paths[fmt.Sprintf("/%s/%s/{%s}", entityName, ActionReferencedBy, pathParameterName)] = map[string]interface{}{
			"get": map[string]interface{}{
				"parameters": []interface{}{
					pathParameter,
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "The references to " + entityName,
						"schema": map[string]interface{}{
							"$ref": "#/definitions/" + referencedByEntityName,
						},
					},
				},
			},
		}

		definition, err := createSwaggerDefinition(entity.New().Collapse())
		if err != nil {
			return "", err
		}
		definitions[entityName] = definition

		expandedDefinition, err := createSwaggerDefinition(entity.New())
		if err != nil {
			return "", err
		}
		definitions[expandedEntityName] = expandedDefinition

		referencedByMap, err := entities.CreateReferencedByMap(entityName)
		if err != nil {
			return "", err
		}

		referencedByDefinition, err := createSwaggerDefinition(referencedByMap)
		if err != nil {
			return "", err
		}
		definitions[referencedByEntityName] = referencedByDefinition
	}

	swagger := map[string]interface{}{
		"swagger": "2.0",
		"info": map[string]interface{}{
			"version": info.Version,
			"title":   info.Title,
		},
		"host":        host,
		"paths":       paths,
		"definitions": definitions,
	}

	content, err := json.Marshal(swagger)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// TODO implement
func createSwaggerDefinition(in interface{}) (interface{}, error) {
	return map[string]interface{}{"type": "string"}, nil
}

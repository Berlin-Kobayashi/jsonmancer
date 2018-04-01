package storage

import (
	"encoding/json"
	"fmt"
	"reflect"
	"bytes"
	"errors"
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

		definition, err := CreateSwaggerDefinition(entity.New().Collapse())
		if err != nil {
			return "", err
		}
		definitions[entityName] = definition

		expandedDefinition, err := CreateSwaggerDefinitionForResource(entity)
		if err != nil {
			return "", err
		}
		definitions[expandedEntityName] = expandedDefinition

		referencedByMap, err := entities.CreateReferencedByMap(entityName)
		if err != nil {
			return "", err
		}

		referencedByDefinition, err := CreateSwaggerDefinition(referencedByMap)
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

func CreateSwaggerDefinitionForResource(in Entity) (interface{}, error) {
	data, err := CreateSwaggerDefinition(reflect.New(in.Data).Interface())
	if err != nil {
		return nil, err
	}

	references := map[string]interface{}{}
	for relationName, reference := range in.References {
		referenceSwagger, err := CreateSwaggerDefinitionForResource(reference)
		if err != nil {
			return nil, err
		}

		references[toLowerFirstLetter(relationName)] = referenceSwagger
	}

	properties := map[string]interface{}{
		"id":         map[string]interface{}{"type": "string"},
		"data":       data,
		"references": map[string]interface{}{"type": "object", "properties": references},
	}

	return map[string]interface{}{"type": "object", "properties": properties}, nil
}

func CreateSwaggerDefinition(in interface{}) (interface{}, error) {
	if in == nil {
		return nil, errors.New("cannot create swagger definition for nil")
	}

	t := reflect.TypeOf(in)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	v := reflect.ValueOf(in)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]interface{}{"type": "integer"}, nil
	case reflect.Float32, reflect.Float64:
		return map[string]interface{}{"type": "number"}, nil
	case reflect.Bool:
		return map[string]interface{}{"type": "boolean"}, nil
	case reflect.String:
		return map[string]interface{}{"type": "string"}, nil
	case reflect.Slice:
		items, err := CreateSwaggerDefinition(reflect.New(t.Elem()).Interface())
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"type": "array", "items": items}, nil
	case reflect.Struct:
		properties := map[string]interface{}{}
		for i := 0; i < t.NumField(); i++ {
			fieldName := t.Field(i).Name
			if startsWithCapitalLetter(fieldName) {
				field, err := CreateSwaggerDefinition(v.FieldByName(fieldName).Interface())
				if err != nil {
					return nil, fmt.Errorf("struct field %q: %q", fieldName, err.Error())
				}

				properties[toLowerFirstLetter(fieldName)] = field
			}
		}

		return map[string]interface{}{"type": "object", "properties": properties}, nil
	case reflect.Map:
		properties := map[string]interface{}{}

		if t.Key().Kind() != reflect.String {
			return nil, fmt.Errorf("unsupported map key type %q", t.Kind())
		}

		for _, k := range v.MapKeys() {
			field, err := CreateSwaggerDefinition(v.MapIndex(k).Interface())
			if err != nil {
				return nil, fmt.Errorf("map key %q: %q", k.String(), err.Error())
			}

			properties[toLowerFirstLetter(k.String())] = field
		}

		return map[string]interface{}{"type": "object", "properties": properties}, nil
	}

	return nil, fmt.Errorf("unsupported data type %q", t.Kind())
}

func startsWithCapitalLetter(s string) bool {
	return s != toLowerFirstLetter(s)
}

func toLowerFirstLetter(in string) string {
	if in == "" {
		return in
	}

	out := []byte(in)
	out[0] = bytes.ToLower([]byte{out[0]})[0]

	return string(out)
}

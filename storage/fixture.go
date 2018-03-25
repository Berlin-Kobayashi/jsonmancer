package storage

import "reflect"

type FixtureDataType struct {
	Data string
	Nested struct {
		Data string
	}
}

var fixtureStorage, _ = New(FixtureEntities, dummyRepository{}, dummyUUIDGenerator{})

var FixtureEntities = []Entity{fixtureReferencingEntity, fixtureReferencedEntity}

var fixtureReferencingEntity = Entity{
	Name:       fixtureReferencingEntityName,
	Data:       reflect.TypeOf(FixtureDataType{}),
	References: map[string]Entity{"reference": fixtureReferencedEntity},
}

var fixtureReferencedEntity = Entity{
	Name: fixtureReferencedEntityName,
	Data: reflect.TypeOf(FixtureDataType{}),
}

var FixtureReferencingResource = Resource{
	ID:         fixtureReferencingID,
	Data:       fixtureReferencingData,
	References: fixtureReferences,
}

var fixtureReferencingID = "1"
var fixtureReferencingData = FixtureDataType{
	Data:   "referencingData",
	Nested: struct{ Data string }{Data: "referencingNestedData"},
}

var fixtureReferences = map[string][]Resource{
	"reference": {fixtureReferencedResource},
}

var fixtureReferencedResource = Resource{
	ID:   fixtureReferencedID,
	Data: fixtureReferencedData,
}

var fixtureReferencingEntityName = "referencingEntity"
var fixtureReferencedEntityName = "referencedEntity"

var fixtureReferencedID = "2"
var fixtureReferencedData = FixtureDataType{
	Data:   "referencedData",
	Nested: struct{ Data string }{Data: "referencedNestedData"},
}

var uuidV4Fixture = "b5e57615-0f40-404e-bbe0-6ae81fe8080a"

var missingIDFixture = "123"

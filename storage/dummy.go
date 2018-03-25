package storage

type dummyUUIDGenerator struct {
}

func (g dummyUUIDGenerator) Generate() string {
	return uuidV4Fixture
}

var savedData interface{}
var updatedData interface{}
var deletedData []string
var queriedData Query

type dummyRepository struct {
}

func (s dummyRepository) Create(collectionName string, data interface{}) error {
	savedData = data

	return nil
}

func (s dummyRepository) Read(collectionName string, id string, result *interface{}) error {
	if id == missingIDFixture {
		return NotFound{}
	}

	switch collectionName {
	case fixtureReferencingEntityName:
		*result = FixtureReferencingResource.Collapse()
	case fixtureReferencedEntityName:
		*result = fixtureReferencedResource.Collapse()
	default:
		return NotFound{}
	}

	return nil
}

func (s dummyRepository) Update(collectionName string, id string, data interface{}) error {
	updatedData = data

	return nil
}

func (s dummyRepository) Delete(collectionName string, id string) error {
	if id == missingIDFixture {
		return NotFound{}
	}

	deletedData = append(deletedData, id)

	return nil
}

func (s dummyRepository) ReadAll(collectionName string, query Query, result *[]interface{}) error {
	queriedData = query

	var data interface{}

	err := s.Read(collectionName, "", &data)
	if err != nil {
		return err
	}

	*result = []interface{}{data}

	return nil
}

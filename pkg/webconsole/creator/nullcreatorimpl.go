package creator

type NullCreatorImpl struct {
}

func (con *NullCreatorImpl) Create(yamlStr string) (bool, error) {
	//put your logic here
	return false, nil;
}

func NewNullCreatorImpl() Creator {
	return new(NullCreatorImpl)
}

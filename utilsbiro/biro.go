package utilsbiro

import (
	"reflect"
)

type TypeGenericSourceFinder func(id int) (interface{}, bool, error)

func (TypeGenericSourceFinder) IsValidResults(in []reflect.Value, out []reflect.Value) bool {
	return IsValidResults_ForFindBiroLegacy(in, out)
}

//generic result validator for biro findings
func IsValidResults_ForFindBiroLegacy(in []reflect.Value, out []reflect.Value) bool {
	return out[1].Bool() //cache if has some returned value
}

//call generic source finder and transform result to biro result, ItemV1, ItemOfferV1 etc...
type TransformerFinder func(id int, params map[string]string, version string, finder TypeGenericSourceFinder) (interface{}, bool, error)

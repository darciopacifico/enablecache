package cache

import (
	"encoding/gob"
	"errors"
	"fmt"
	"math"
	"reflect"
)

//
type AutoCacheManager struct {
	Ps CacheStorage
}

const TTL_INFINITY = -1
const CacheTreeRefPrefix = "attTreeRef"

//type MapVisits map[string]CacheRegistry //-> contains a map of cacheKey->value

//Mark    struct as a cacheable entity. Allow struct to generate  key about itself

type Cacheable interface {
	GetCacheKey() string
}

//register map[string]interface {} in gob
func init() {
	gob.Register(make(map[string]interface{}, 0))
}

//invalidate cache registry
func (c AutoCacheManager) Invalidate(cacheKeys ...string) error {

	errDel := c.Ps.DeleteValues(cacheKeys...)

	if errDel != nil {
		log.Error("Error trying to delete values from cache %v", errDel)
	}

	return errDel
}

func (c AutoCacheManager) Validade() bool {
	return true
}

//set cache implementation
func (c AutoCacheManager) SetCache(cacheRegistries ...CacheRegistry) error {

	mapToSave := make(map[string]CacheRegistry)

	for _, cacheRegistry := range cacheRegistries {

		_, err := c.mapAttributesToCacheKeys(cacheRegistry.CacheKey, &cacheRegistry, &mapToSave)
		if err != nil {
			return err
		}
	}

	//return nil
	return c.Ps.SetValues(c.mapToArray(mapToSave)...)
}

//convert the map to array
func (c AutoCacheManager) mapToArray(mapToSave map[string]CacheRegistry) []CacheRegistry {

	arrCacheRegistry := make([]CacheRegistry, 0)

	for _, cacheRegistry := range mapToSave {
		arrCacheRegistry = append(arrCacheRegistry, cacheRegistry)
	}

	return arrCacheRegistry
}

//return time to live
func (c AutoCacheManager) GetCacheTTL(cacheKey string) (int, error) {
	return c.Ps.GetTTL(cacheKey)
}

//get cache for only one cachekey
func (c AutoCacheManager) GetCache(cacheKey string) (CacheRegistry, error) {
	cacheRegs, err := c.GetCaches(cacheKey)
	if err != nil {
		return CacheRegistry{cacheKey, nil, -2, false}, err
	}

	if len(cacheRegs) > 0 {
		return cacheRegs[cacheKey], nil
	} else {
		return CacheRegistry{cacheKey, nil, -2, false}, nil
	}
}

//implement getCache operation that can recover child data in other cache registries.
func (c AutoCacheManager) GetCaches(cacheKeys ...string) (map[string]CacheRegistry, error) {

	//recover keys for auto buildUp.
	// actually just make a copy of cacheKeys array and append a prefix on the values
	ckDepTree := c.getCKDepTree(cacheKeys...)

	//recover cache regs for build tree. ask cache storage
	crDepTree, err := c.Ps.GetValuesMap(ckDepTree...)
	if err != nil {
		return make(map[string]CacheRegistry, 0), err
	}

	//navigate over tree maps and accumulate all required cacheKeys
	dependencyKeys := getDependencyKeys(crDepTree)

	//aggregate cacheKeys and additionalKeys,
	// to make only one roundtrip do cachestorage
	allKeys := append(dependencyKeys, cacheKeys...)

	//recover cacheregs for all keys, in ony one round trip
	allCRs, err := c.Ps.GetValuesMap(allKeys...)

	//check cache miss
	//all the informed cacheKeys must be returned by cache storage
	if c.hasSomeCacheMiss(cacheKeys, allCRs) {
		log.Debug("one or many kes was missed. Required keys: %v", cacheKeys)
		return make(map[string]CacheRegistry, 0), nil
	}

	//iterate over cacheKeys and build up the registries, based on tree of references, crTreeRefs
	mapCR, err := c.buildUpCRs(cacheKeys, crDepTree, &allCRs)
	if err != nil {
		log.Debug("Error trying to build up values for keys %v", cacheKeys)
		return make(map[string]CacheRegistry, 0), err
	}

	return mapCR, nil
}

//take all cache registries returned and build them up
func (c AutoCacheManager) buildUpCRs(cacheKeys []string, crTreeRefs map[string]CacheRegistry, allCRs *map[string]CacheRegistry) (map[string]CacheRegistry, error) {

	var mountedCRs = make(map[string]CacheRegistry, len(cacheKeys)) //make the final map

	for _, cacheKey := range cacheKeys { //iterate over requested keys

		cacheRegistry := (*allCRs)[cacheKey]                              //cache registry. No presence check is needed. hasSomeCachemiss already check that...
		crTreeRef, hasTree := crTreeRefs[getKeyForDependencies(cacheKey)] //tree of values to rebuild the cache registry

		if hasTree && crTreeRef.HasValue { // check whether cache key has a tree of attributes and cachevalues to build up the cacheRegistry
			err := c.buildUpCR(&cacheRegistry, crTreeRef, allCRs)

			if err != nil { //at this time, no error will be accepted.
				log.Error("One or many cache registries was missed! %v", err)
				//any error invalidate all cache operation
				return make(map[string]CacheRegistry, 0), err
			}
		}

		mountedCRs[cacheKey] = cacheRegistry //acumulate results in final map

	}

	return mountedCRs, nil // voila!

}

//just wrap cache registry in a reflect.Value, cast tree payload to a map and call recursive build up
func (c AutoCacheManager) buildUpCR(cacheRegistry *CacheRegistry, crTreeRef CacheRegistry, allCRs *map[string]CacheRegistry) error {

	//type assert to tree of values (a map actually)
	mapAttributes := crTreeRef.Payload.(map[string]interface{})

	//call a reflexive builup
	value, err, ttl := c.buildUp(cacheRegistry, mapAttributes, allCRs)
	if err != nil {
		return err
	}

	cacheRegistry.Ttl = ttl
	cacheRegistry.Payload = value.Interface()

	//is everything alright
	return nil

}

func getDependencyKeys(crTreeRefs map[string]CacheRegistry) []string {

	additionalKeys := make([]string, 0)
	for _, crTreeRef := range crTreeRefs {
		if crTreeRef.HasValue {

			//must be a map. Nothing more is expected
			mapTreeRef := crTreeRef.Payload.(map[string]interface{})

			//navigate over tree of references map, identify all cache keys needed
			dependencyKeys := recoverDependencyKeys(mapTreeRef)

			//fmt.Println("Árvore de Referências para", crTreeRef.CacheKey)
			//printKeys(mapTreeRef, "    ")

			//append to additional Keys
			additionalKeys = append(additionalKeys, dependencyKeys...)
		}
	}

	return additionalKeys
}

//recover additional keys from map
func recoverDependencyKeys(crTreeRefs map[string]interface{}) []string {
	arrAddKeys := make([]string, 0)

	accumulateKeys(crTreeRefs, &arrAddKeys)

	return arrAddKeys
}

//navigate over the mapDependencies (tree) and accumulate all cache keys that must be returned
func accumulateKeys(mapDependencies map[string]interface{}, additionalKeys *[]string) {
	for _, itfMapCKs := range mapDependencies { //iterate over attributes

		mapCKs := itfMapCKs.(map[string]interface{})

		for cachekey, itfMapAtt2 := range mapCKs { //iterate over cacheKeys for attributes

			(*additionalKeys) = append((*additionalKeys), cachekey) // accumulate keys

			mapAtt2, _ := itfMapAtt2.(map[string]interface{})

			accumulateKeys(mapAtt2, additionalKeys) //recursive call to next level

		}
	}
}

//navigate over the mapDependencies (tree) and accumulate all cache keys that must be returned
func printKeys(mapDependencies map[string]interface{}, ident string) {

	for attName, itfMapCKs := range mapDependencies { //iterate over attributes

		mapCKs := itfMapCKs.(map[string]interface{})

		for cachekey, itfMapAtt2 := range mapCKs { //iterate over cacheKeys for attributes

			fmt.Println(ident, "att ", attName, "cacheKey ", cachekey)

			mapAtt2, _ := itfMapAtt2.(map[string]interface{})

			printKeys(mapAtt2, ident+ident) //recursive call to next level

		}
	}
}

//return aditional keys for auto build up operation
func (c AutoCacheManager) getCKDepTree(cacheKeys ...string) []string {

	dependencyKeys := make([]string, len(cacheKeys))

	for index, cacheKey := range cacheKeys {
		dependencyKeys[index] = getKeyForDependencies(cacheKey)
	}

	return dependencyKeys
}

//Based on cacheManager algorithm implementation, calculates the final ttl of requested value
func (c AutoCacheManager) calculateTTL(cacheRegistry CacheRegistry, mapBuildUp map[string]CacheRegistry) CacheRegistry {

	//max of ttl type. if still equals maxint32 at final, will be changed to -1, meaning infinite ttl
	ttl := math.MaxInt32

	//range over all child values, looking for smaller ttl, except for -1, that means infinite
	for _, payload := range mapBuildUp {
		if payload.GetTTL() != -1 && payload.GetTTL() < ttl {
			ttl = payload.GetTTL()
		}
	}

	//compare the smaller ttl, except for -1, that means infinite
	if cacheRegistry.GetTTL() != -1 && cacheRegistry.GetTTL() < ttl {
		ttl = cacheRegistry.GetTTL()
	}

	//there is no ttl defined, remaining equals maxint32.
	if ttl == math.MaxInt32 {
		ttl = -1 //infinite
	}

	//update ttl attribute
	cacheRegistry.Ttl = ttl

	return cacheRegistry

}

//test whether all requested keys has returned
func (c AutoCacheManager) hasSomeCacheMiss(keys []string, mapCheckReturn map[string]CacheRegistry) bool {

	if len(keys) > len(mapCheckReturn) { // if the number of asked keys is greater than returned values, meant that some val was missed
		return true
	}

	//iterate returm map, checking hasValue confirmation
	for _, key := range keys {

		cacheRegistry := mapCheckReturn[key]

		if cacheRegistry.Payload == nil || !cacheRegistry.HasValue {
			log.Debug("nao foi encontrado o valor para a chave ", key, "invalidando todo o cache!", cacheRegistry.Payload)
			return true //means cache miss :(
		}
	}

	return false //means cache hit :)
}

//return whether the cacheKey has been visited. if not, mark as visited also
func (c AutoCacheManager) visited(visiteds map[string]CacheRegistry, cacheRegistry CacheRegistry) bool {
	//check whether visited before
	_, hasVisited := visiteds[cacheRegistry.CacheKey]
	return hasVisited
}

//recursivelly rebuild the value, based on tree of attributes and values.
func (c AutoCacheManager) buildUp(cacheRegistry *CacheRegistry, mapAttributes map[string]interface{}, mapVisits *map[string]CacheRegistry) (reflect.Value, error, int) {

	//create a reflect.Value representation for cacheRegistry.Payload, fit and ready for reflect operations
	payloadValue := getValueForPayload(cacheRegistry)

	// iterate over all attributes(aka fields), based on mapAttributes (must be compatible with fields value. No check for that...)
	for attributeName, interfaceAtt := range mapAttributes {
		//cast to map
		mapCacheKeys := interfaceAtt.(map[string]interface{})

		//get attribute from value, based o attribute name
		attribute := payloadValue.FieldByName(attributeName)

		//iterate over mapCachekeys, identifying the values to be put in the attribute, as single value or appended as array..
		for cacheKey, interfaceAtt2 := range mapCacheKeys {
			mapAttributes2 := interfaceAtt2.(map[string]interface{})

			//recover the object to put in the attribute/field, as single value or appended as array..
			crCK, hasVal := (*mapVisits)[cacheKey]
			if !hasVal {
				return payloadValue, errors.New(fmt.Sprintf("Cachekey %v to put in the attribyte %v was not found!", cacheKey, attributeName)), -2
			}

			//recursivelly buildup the cacheValue, as same way of value
			buildedUpCacheValue, err, _ := c.buildUp(&crCK, mapAttributes2, mapVisits)
			//buildedUpCacheValue, err := c.buildUp(&crCK, mapAttributes2, mapVisits)
			if err != nil { //formal check for error
				log.Error("Error trying to build up attribute %v, setting or appending value %v", attributeName, cacheKey)
				return payloadValue, err, -2
			}

			//check the kind of attribute (array or single value)
			//set or append the buildedUpCacheValue to the attribute
			setAttribute(&attribute, buildedUpCacheValue)

			//check the ttl of parent and child and set the lower one to parent always
			setTTL(cacheRegistry, &payloadValue, &crCK, &buildedUpCacheValue)

			//log.Warning("ttl for registry %v %v ", cacheRegistry.CacheKey, cacheRegistry.Ttl)
		}
	}

	//renew the cacheregistry payload attribute with a recently builded up interface that came from payloadValue
	cacheRegistry.Payload = payloadValue.Interface()

	cacheRegistry.Ttl = cacheRegistry.Ttl - 50

	//finally, return the builded up value
	return payloadValue, nil, cacheRegistry.Ttl
}

//check the ttl of parent and child and set the lower one to parent always
func setTTL(crParent *CacheRegistry, valParent *reflect.Value, crChild *CacheRegistry, valChild *reflect.Value) {

	//og.Error("Setando ttl pai %v, ttl filho %v ", crParent.Ttl, crChild.Ttl)

	crParent.Ttl = MinTTL(crParent.Ttl, crChild.Ttl)

	//log.Error("##menor ttl %v %v", crParent.CacheKey, crParent.Ttl)
	setTTLToPayload(crParent)

	//payload := crParent.Payload
	//exposeTTL, hasTtl := payload.(ExposeTTL)

	//if hasTtl {
	//	log.Error("&&menor ttl %v ", exposeTTL.GetTtl())
	//}
	//log.Error("**menor ttl %v %v", crParent.CacheKey, crParent.Ttl)
}

//crate a reflect.Value for cacheRegistry.Payload
func getValueForPayload(cacheRegistry *CacheRegistry) reflect.Value {
	//determine the type of payload
	payload := cacheRegistry.Payload
	payloadVal := reflect.ValueOf(payload)
	payloadType := payloadVal.Type()

	//instantiate a new value of that type
	value := reflect.New(payloadType).Elem()

	//set the value of old payload to the new one
	value.Set(payloadVal)

	return value
}

//if attribute is a single value, set the cacheValue to the attribute
//if attribute is an array, append the cacheValue to the attribute
func setAttribute(attribute *reflect.Value, cacheValue reflect.Value) {

	//check the kind of attribute (array or single value)
	if (*attribute).Kind() == reflect.Array || (*attribute).Kind() == reflect.Slice {
		(*attribute).Set(reflect.Append((*attribute), cacheValue)) //just append the value, like an array

	} else {
		//set the final cachevalue to sttribute
		(*attribute).Set(cacheValue) // set the value as a single value
	}
}

//get the element type for an array or get the type for a single value.
//because all child values must be stored and recovered individually
func typeForAttribute(attribute reflect.Value) reflect.Type {
	var typeForNewVal reflect.Type
	if attribute.Kind() == reflect.Array || attribute.Kind() == reflect.Slice {
		typeForNewVal = attribute.Type().Elem() // type for array element
	} else {
		typeForNewVal = attribute.Type() // type for the attribute, as single value
	}

	//type for attribute
	return typeForNewVal
}

//range over object attributes, determining which cachekey each attribute contains
//return a map of attributes->cacheKeys->attribues->cachekeys ... and so on, and fill mapVisits
func (c AutoCacheManager) mapAttributesToCacheKeys(cacheKey string, cacheRegistry *CacheRegistry, mapVisits *map[string]CacheRegistry) (map[string]interface{}, error) {

	mapAttribute := make(map[string]interface{})
	//check whether already visited this value before...

	if _, hasVal := (*mapVisits)[cacheKey]; hasVal {
		//OK, ALREADY VISITED. BREAKING LOOP, GETTING OUT...
		return mapAttribute, nil
	}

	//MARK AS VISITED, including in visiteds map
	(*mapVisits)[cacheKey] = (*cacheRegistry)

	//wrap cacheregistry.payload in a reflect.value for reflect operations
	value := getValueForPayload(cacheRegistry)

	// counting attributes and creating map containing the cacheKey mapped to each attribute of passed value
	numAttributes := value.NumField()

	//Iterate over attributes of value
	for i := 0; i < numAttributes; i++ {

		//get the field value
		field := value.Field(i)

		//recover the values associated to the attribute.
		//if is an array, return all values separatelly
		//if is an single value, return one item array(len=1)
		fieldValues := getFieldValues(field)

		//make a map to store the cache keys
		mapCacheKey := make(map[string]interface{})

		//iterate over discovered values of the field
		for _, fieldValue := range fieldValues {
			_ = fieldValue

			//discover the cache key value for this value
			cacheKey, err := GetKey(fieldValue)
			if err != nil {
				return nil, err
			}

			//recursivelly inspect this fieldValue as same as value passed to this function before
			cacheRegistryField := CacheRegistry{cacheKey, fieldValue.Interface(), cacheRegistry.Ttl, true}

			//mapAttributes, err := c.MapAttributesToCacheKeys(cacheKey, fieldValue, mapVisits)
			mapAttributes, err := c.mapAttributesToCacheKeys(cacheKey, &cacheRegistryField, mapVisits)

			if err != nil {
				return nil, err
			}

			//in mapcachekey, store the map of attributes per cachekey
			mapCacheKey[cacheKey] = mapAttributes

		}

		//if no value was identified, there is nothing to save
		if len(mapCacheKey) > 0 {
			fieldName := value.Type().Field(i).Name
			mapAttribute[fieldName] = mapCacheKey

			//clear the attribute
			typeForAttribute := field.Type()

			newEmptyVal := reflect.New(typeForAttribute).Elem()
			field.Set(newEmptyVal)

			//log.Error("TIPO DO ATRIBUTO %v valor %v", field.Type(), field.Interface())

		}
	}

	//put the dried value on map visits
	cacheRegistry.Payload = value.Interface()
	//(*mapVisits)[cacheKey] = CacheRegistry{cacheKey, value.Interface(), -1, true}

	(*mapVisits)[cacheKey] = (*cacheRegistry)

	if len(mapAttribute) > 0 {
		//put the tree of references
		treeCacheKey := getKeyForDependencies(cacheKey)
		(*mapVisits)[treeCacheKey] = CacheRegistry{treeCacheKey, mapAttribute, -1, true}
	}

	//return all stuff
	return mapAttribute, nil

}

func getKeyForDependencies(cacheKey string) string {
	return cacheKey + ":" + CacheTreeRefPrefix
}

//iterate over field values, if is an array, or simply get value, if is not an array
func getFieldValues(field reflect.Value) []reflect.Value {
	var values []reflect.Value

	if field.IsValid() && (field.Kind() == reflect.Array || field.Kind() == reflect.Slice) && field.Cap() > 0 {
		qtd := field.Len()

		log.Debug("Field is an array, size %v. Trying to split this array value in multiple reflect.Value's!", field.Len())

		for i := 0; i < qtd; i++ {
			//log.Debug("Checking cache for index %v", i)
			addIfCacheable(&values, field.Index(i))
		}
	} else {
		addIfCacheable(&values, field)
	}

	return values

}

//only cacheable values will be inspected at attribute level and, perhaps, stored separately.
//non cacheable values, will be stored at all
func addIfCacheable(values *[]reflect.Value, value reflect.Value) {
	if value.IsValid() && value.CanInterface() { // i really try to put these two ifs at the same line... sorry about that :-(
		if cacheable, isCacheable := value.Interface().(Cacheable); isCacheable && len(cacheable.GetCacheKey()) > 0 {
			//the value is cacheable and key is not empty
			*values = append(*values, value)
		}
	}
}

//Retrieve the cachekey for some object
//Error if object is not Cacheable AND is not possible to inter cacheKey
func GetKey(value reflect.Value) (string, error) {

	object := value.Interface()

	cacheable, isCacheable := object.(Cacheable)

	if isCacheable {
		strCacheKey := cacheable.GetCacheKey()
		return strCacheKey, nil
	} else {
		strId, err := GetIdFieldStr(object)

		if err != nil {
			return "", err
		} else {
			strCacheKey := reflect.TypeOf(object).Name() + ":" + strId
			return strCacheKey, nil
		}
	}
}

//Return the "Id" field, if it exists, as a string val
//Error if object does't have an Id field
func GetIdFieldStr(object interface{}) (string, error) {

	if object == nil {
		return "", errors.New("Object is nil!")
	}

	value := reflect.ValueOf(object)

	if value.IsValid() {
		fieldId := value.FieldByName("Id")

		if fieldId.IsValid() {
			idAsInterface := fieldId.Interface()
			isAsStr := fmt.Sprintf("%v", idAsInterface)
			return isAsStr, nil
		} else {
			return "", errors.New("Object has no 'Id' field! Implement 'Cacheable' interface for cache operations!")
		}

	} else {

		return "", errors.New("Object is not valid!")
	}
}

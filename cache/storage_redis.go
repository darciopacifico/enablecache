package cache

import (
	//"bytes"
	//"encoding/gob"
	"time"

	"github.com/garyburd/redigo/redis"
	"errors"
)

//Cache storage implementation using redis as key/value storage
type RedisCacheStorage struct {
	redisPool      	redis.Pool
	ttlReadTimeout 	int
	cacheArea      	string
	enableTtl	 	bool
	Serializer     	Serializer // usually SerializerGOB implementation
}

var _=SerializerGOB{} // this is the usual serializer used above!!



//recover all cacheregistries of keys
func (s RedisCacheStorage) GetValuesMap(cacheKeys ...string) (mapResp map[string]CacheRegistry, retError error) {

	defer func() { //assure for not panicking
		if r := recover(); r != nil {
			log.Error("Recovering error from Redis Cache Storage!!  %v", r)
			log.Error("Returning as no cached registry found!!")

			mapResp = make(map[string]CacheRegistry)
			retError = errors.New("Error trying to get values map")
			return
		}
	}()

	ttlMapChan := make(chan map[string]int, 1)

	//if (s.enableTtl) {
	if (false) { // param does not correct. not important now
		go func() { // parallel ttl recover
			defer func() {
				if r := recover(); r != nil {
					log.Critical("Error trying to get ttl for registries %v!", cacheKeys)

					//in case of error, retur an empty map
					ttlMapChan <- make(map[string]int, 0)
				}
			}()

				//put result on channel
				ttlMapChan <- s.GetTTLMap(cacheKeys)
		}()
	}

	mapCacheRegistry := make(map[string]CacheRegistry)

	if len(cacheKeys) <= 0 {
		log.Debug("Nenhuma chave informada para busca. len(arrKeys)=0!")
		return mapCacheRegistry, nil //empty map
	}

	conn := s.redisPool.Get()
	if(conn==nil){
		log.Error("Error trying to acquire redis conn! null connection")
		return make(map[string]CacheRegistry), errors.New("Redis conn is null! Check conn errors!")
	}

	defer conn.Close()
	var err error = nil

	//log.Debug(cacheKeys)

	replyMget, err := conn.Do("MGET", (s.getKeys(cacheKeys))...)
	if err != nil || replyMget == nil {
		log.Error("Error trying to get values from cache %v", err)
		log.Error("Returning an empty registry!")

		return mapCacheRegistry, err // error trying to search cache keys
	}

	arrResults, isArray := replyMget.([]interface{}) //try to convert the returned value to array

	if !isArray {
		log.Error("Value returned by a MGET query is not array for keys %v! No error will be returned!", cacheKeys) //formal check
		return make(map[string]CacheRegistry), nil
	}

	for _, cacheRegistryNotBytes := range arrResults {
		if cacheRegistryNotBytes != nil {


/*
			cacheRegistryBytes, isByteArr := cacheRegistryNotBytes.(string)
			if(isByteArr){
				log.Error("error trying to deserialize! not a byte array")
				return mapCacheRegistry, errors.New("not byte array!")
			}
*/


			cacheRegistryBytes, errBytes := redis.Bytes(cacheRegistryNotBytes, err)
			if errBytes != nil || replyMget == nil {
				return mapCacheRegistry, errBytes
			}

			cacheRegistry := CacheRegistry{}

			interfaceResp, _, errUnm := s.Serializer.UnmarshalMsg(cacheRegistry,cacheRegistryBytes)
			if errUnm!=nil {
				log.Error("error trying to deserialize!",errUnm)
				return mapCacheRegistry, errUnm
			}

			cacheRegistry, isCR := interfaceResp.(CacheRegistry)
			if(!isCR){
				log.Error("error trying to deserialize! object is not a CacheRegistry object type!")
				return mapCacheRegistry, nil
			}

			if err != nil {
				log.Error("Warning!! Error trying to recover data from redis!", err)
			} else {
				if cacheRegistry.Payload == nil {
					log.Error("ATENCAO! NENHUM PAYLOAD FOI RETORNADO DO REDIS!")
				}
				//Everything is alright
				mapCacheRegistry[cacheRegistry.CacheKey] = cacheRegistry
			}
		}
	}

	//if (s.enableTtl) {
	if (false) { // error returning param. not important now
		select {
		//wait for ttl channel
		case ttlMap := <-ttlMapChan:
			mapCacheRegistry = s.zipTTL(mapCacheRegistry, ttlMap)
		//in case of timeout, returt an empty map
		case <-time.After(time.Duration(s.ttlReadTimeout) * time.Millisecond):
			log.Warning("Retrieve TTL for cachekeys %v from redis timeout after %dms, continuing without it.", cacheKeys, s.ttlReadTimeout)
			mapCacheRegistry = s.zipTTL(mapCacheRegistry, make(map[string]int, 0))
		}
	}

	return mapCacheRegistry, nil // err=nil by default, if everything is alright
}

//Recover current ttl information about registry
func (s RedisCacheStorage) GetTTL(key string) (int, error) {
	oneItemMap := make(map[string]CacheRegistry, 1)

	oneItemMap[key] = CacheRegistry{key, "", -2 /*not found*/, true, ""}

	respMap, errTTL := s.GetActualTTL(oneItemMap)
	return respMap[key].Ttl, errTTL

}

//Recover current ttl information about registries
func (s RedisCacheStorage) zipTTL(mapCacheRegistry map[string]CacheRegistry, ttlMap map[string]int) map[string]CacheRegistry {
	//prepare a keyval pair array
	for key, cacheRegistry := range mapCacheRegistry {
		if ttl, hasTtl := ttlMap[key]; hasTtl {
			cacheRegistry.Ttl = ttl
		} else {
			cacheRegistry.Ttl = -1
		}
		mapCacheRegistry[key] = cacheRegistry
	}

	return mapCacheRegistry
}

//Recover current ttl information about registries
func (s RedisCacheStorage) GetActualTTL(mapCacheRegistry map[string]CacheRegistry) (returnMap map[string]CacheRegistry, retError error) {

	defer func() { //assure for not panicking
		if r := recover(); r != nil {
			log.Error("TTL Recovering error from Redis Cache Storage!!  %v", r)
			log.Error("Returning as no TTL info found!!")

			returnMap = mapCacheRegistry
			retError = errors.New("Error trying to get actual ttl val!")

			return
		}
	}()


	conn := s.redisPool.Get()
	if(conn==nil){
		log.Error("TTL: Error trying to acquire redis conn! null connection")
		return make(map[string]CacheRegistry), errors.New("TTL: Redis conn is null! Check conn errors!")
	}
	defer conn.Close()

	//prepare a keyval pair array
	for keyMap, cacheRegistry := range mapCacheRegistry {

		respTtl, err := conn.Do("ttl", s.getKey(keyMap))
		log.Debug("TTL %v that came from redis %v", keyMap, respTtl)

		if err != nil {
			log.Error("Error trying to retrieve ttl of key " + keyMap, err)
			cacheRegistry.Ttl = -2
			return mapCacheRegistry, err

		} else {
			intResp, _ := respTtl.(int64)
			cacheRegistry.Ttl = int(intResp)
		}

		mapCacheRegistry[keyMap] = setTTLToPayload(&cacheRegistry)
	}

	return mapCacheRegistry, nil
}

//Recover current ttl information about registries
func (s RedisCacheStorage) GetTTLMap(keys []string)  (retTTLMap map[string]int ){

	defer func() { //assure for not panicking
		if r := recover(); r != nil {
			log.Error("TTL error from Redis Cache Storage!!  %v", r)
			log.Error("Returning as emptu ttl map !!")

			retTTLMap = make(map[string]int, 0)
			return
		}
	}()

	ttlMap := make(map[string]int, len(keys))

	conn := s.redisPool.Get()
	if(conn==nil){
		log.Error("TTLMap: Error trying to acquire redis conn! null connection")
		return make(map[string]int)
	}
	defer conn.Close()

	//prepare a keyval pair array
	for _, key := range keys {

		respTtl, err := conn.Do("ttl", s.getKey(key))
		log.Debug("TTL %v that came from redis %v", key, respTtl)

		if err != nil {
			log.Error("Error trying to retrieve ttl of key " + key, err)
			ttlMap[key] = -2

		} else {
			intResp, _ := respTtl.(int64)
			ttlMap[key] = int(intResp)
		}

	}

	return ttlMap
}

//transfer the ttl information from cacheRegistry to paylaod interface, if it is ExposeTTL
func setTTLToPayload(cacheRegistry *CacheRegistry) CacheRegistry {

	payload := cacheRegistry.Payload

	exposeTTL, hasTtl := payload.(ExposeTTL)

	if hasTtl {
		log.Debug("Transfering ttl from redis (%d seconds) registry to ttl attribute of object %s", cacheRegistry.Ttl, cacheRegistry.CacheKey)
		payload = exposeTTL.SetTtl(cacheRegistry.Ttl) // assure the same type, from set ttl
		cacheRegistry.Payload = payload
		log.Debug("Setting ttl to %v, ttl value %v", cacheRegistry.CacheKey, exposeTTL.GetTtl())
	} else {
		log.Debug("Payload doesn't ExposeTTL %v", cacheRegistry.CacheKey)
	}

	return *cacheRegistry
}

//save informed registries on redis
func (s RedisCacheStorage) SetValues(registries ...CacheRegistry) (retErr error) {

	defer func() { //assure for not panicking
		if r := recover(); r != nil {
			log.Error("Error trying to save cacheRegs!!  %v", r)
			log.Error("Returning recovered error!")

			retErr = errors.New("Error trying to save cacheReg")
			return
		}
	}()


	var cacheRegistry CacheRegistry
	var index int

	conn := s.redisPool.Get()
	if(conn==nil){
		log.Error("SetValues: Error trying to acquire redis conn! null connection")
		return errors.New("SetValues: Redis conn is null! Check conn errors!")
	}


	defer conn.Close()

	keyValPairs := make([]interface{}, 2 * len(registries))

	//prepare a keyval pair array
	for index, cacheRegistry = range registries {

		if len(cacheRegistry.CacheKey) == 0 {
			log.Error("cache key vazio !!!")
			//panic(errors.New("cache key vazio"))
		}

		var bytes = []byte{}
		bytes, err := s.Serializer.MarshalMsg(cacheRegistry,bytes)
		if(err!=nil){
			return err
		}


		if len(bytes) == 0 {
			log.Error("Error trying to decode value for key %v", cacheRegistry.CacheKey)
		}

		keyValPairs[(index * 2)] = s.getKey(cacheRegistry.CacheKey)
		keyValPairs[(index * 2) + 1] = bytes

	}

	_, errDo := conn.Do("MSET", keyValPairs...)
	if errDo != nil {
		log.Error("Error trying to save registry! %v %v", s.getKey(cacheRegistry.CacheKey), errDo)
		return errDo
	} else {
		log.Debug("Updating cache reg key %v ", s.getKey(cacheRegistry.CacheKey))
	}

	errF := conn.Flush()
	if errF != nil {
		log.Error("Error trying to flush connection! %v", errF)
		return errF
	}
	s.SetExpireTTL(registries...)
	return nil
}

//set defined ttl to the cache registries
func (s RedisCacheStorage) SetExpireTTL(cacheRegistries ...CacheRegistry) {
	defer func() { //assure for not panicking
		if r := recover(); r != nil {
			log.Error("Error trying to set expire ttl!!  %v", r)
			return
		}
	}()

	conn := s.redisPool.Get()
	if(conn==nil){
		log.Error("SetExpires: Error trying to acquire redis conn! null connection")
		return
	}

	defer conn.Close()

	//prepare a keyval pair array
	for _, cacheRegistry := range cacheRegistries {
		if cacheRegistry.GetTTL() > 0 {
			//log.Debug("Setting ttl to key %s ", cacheRegistry.CacheKey)
			_, err := conn.Do("expire", s.getKey(cacheRegistry.CacheKey), cacheRegistry.GetTTL())
			if err != nil {
				log.Error("Error trying to save cache registry w! %v", err)
				return
			}

		} else {
			log.Debug("TTL for %s, ttl=%d will not be setted! ", s.getKey(cacheRegistry.CacheKey), cacheRegistry.GetTTL())
		}
	}

	err := conn.Flush()
	if err != nil {
		log.Error("Error trying to save cache registry z! %v", err)
		return
	}
}

//delete values from redis
func (s RedisCacheStorage) DeleteValues(cacheKeys ...string) ( retErr error) {
	c := s.redisPool.Get()
	if(c==nil){
		log.Error("Delete: Error trying to acquire redis conn! null connection")
		return errors.New("Delete: Redis conn is null! Check conn errors!")
	}

	defer func() { //assure for not panicking
		if r := recover(); r != nil {
			log.Error("Error trying to delete reg!!  %v", r)

			retErr = errors.New("Error trying to delete reg")
			return
		}

		if(c!=nil){
			c.Close()
		}
	}()


	//apply a prefix to cache area
	keys := s.getKeys(cacheKeys)

	reply, err := c.Do("DEL", keys...)
	if err != nil {
		log.Error("Erro ao tentar invalidar registro no cache!", err, reply)
		return err
	}

	return nil
}

//apply a prefix to cache area
func (s RedisCacheStorage) getKey(key string) string {
	var newKey string

	var serPredix = s.Serializer.GetPrefix()

	if len(s.cacheArea) > 0 {
		newKey = s.cacheArea + serPredix + key
	} else {
		newKey = key
	}

	return newKey
}

//apply a prefix to cachearea
func (s RedisCacheStorage) getKeys(keys []string) []interface{} {

	newKeys := make([]interface{}, len(keys))

	for index, key := range keys {
		newKey := s.getKey(key)
		newKeys[index] = newKey
	}

	return newKeys
}

//instantiate a new cachestorage redis
func NewRedisCacheStorage(hostPort string, password string, maxIdle int, readTimeout int, ttlReadTimeout int, cacheArea string, serializer Serializer, enableTTL bool) RedisCacheStorage {

	redisCacheStorage := RedisCacheStorage{
		*newPoolRedis(hostPort, password, maxIdle, readTimeout),
		ttlReadTimeout,
		cacheArea,
		enableTTL,
		serializer,
	}

	return redisCacheStorage
}

//create a redis connection pool
func newPoolRedis(server, password string, maxIdle int, readTimeout int) *redis.Pool {

	return &redis.Pool{
		MaxIdle:     maxIdle,
		IdleTimeout: 240 * time.Second,
		Dial: func() ( retConn redis.Conn, retErr error) {

			defer func() { //assure for not panicking
				if r := recover(); r != nil {
					log.Error("Error open redis conn!!  %v", r)
					log.Error("Retuning error")

					retConn = nil
					retErr = errors.New("Error trying to open redis conn!!")

					return
				}
			}()

			c, err := redis.Dial("tcp", server, redis.DialReadTimeout(time.Duration(readTimeout) * time.Millisecond))
			if err != nil {
				log.Error("Erro ao tentar se conectar ao redis! ", err)
				return nil, err
			}

			return c, err
		},

		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

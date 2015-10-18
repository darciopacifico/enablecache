package cache

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/garyburd/redigo/redis"
)

//Cache storage implementation using redis as key/value storage
type RedisCacheStorage struct {
	redisPool  redis.Pool
	cacheAreaa string
}

//recover all cacheregistries of keys
func (s RedisCacheStorage) GetValuesMap(cacheKeys ...string) (map[string]CacheRegistry, error) {

	ttlMapChan := make(chan map[string]int)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Critical("Error trying to get ttl for registries %v!", cacheKeys)
				//panic(errors.New("panicking"))
				//return no ttl
				ttlMapChan <- make(map[string]int, 0)
			}
		}()
	}()

	mapCacheRegistry := make(map[string]CacheRegistry)

	if len(cacheKeys) <= 0 {
		log.Debug("Nenhuma chave informada para busca. len(arrKeys)=0!")
		return mapCacheRegistry, nil //empty map
	}

	conn := s.redisPool.Get()
	defer conn.Close()
	var err error = nil

	replyMget, err := conn.Do("MGET", (s.getKeys(cacheKeys))...)
	if err != nil || replyMget == nil {
		log.Error("Error trying to get values from cache %v", err)
		log.Error("Returning an empty registry!")

		return mapCacheRegistry, err // error trying to search cache keys
	}

	arrResults, isArray := replyMget.([]interface{}) //try to convert the returned value to array

	if isArray { //formal check
		for _, cacheRegistryNotBytes := range arrResults {
			if cacheRegistryNotBytes != nil {

				cacheRegistryBytes, errBytes := redis.Bytes(cacheRegistryNotBytes, err)
				if errBytes != nil || replyMget == nil {
					return mapCacheRegistry, errBytes
				}

				bufferResp := bytes.NewBuffer(cacheRegistryBytes)

				d := gob.NewDecoder(bufferResp) //instantiate a decoder base on bytes

				cacheRegistry := CacheRegistry{}

				err = d.Decode(&cacheRegistry) // try to decode this bytes in a cacheRegistry object
				if err != nil {
					log.Error("Warning!! Error trying to recover data from redis!", err)
				} else {

					if cacheRegistry.Payload == nil {
						log.Error("ATENCAO! NENHUM PAYLOAD FOI RETORNADO DO REDIS!")
					} else {

						log.Debug("Retornando cache key %v do redis!", cacheRegistry.CacheKey)
					}

					//Everything is alright
					mapCacheRegistry[cacheRegistry.CacheKey] = cacheRegistry
				}
			} else {
				log.Debug("Returned null for cacheKeys %v!", cacheKeys)
			}
		}
	} else {
		log.Error("Value returned by a MGET query is not array for keys %v! No error will be returned!", cacheKeys) //formal check

		return make(map[string]CacheRegistry), nil
	}

	//wait for ttl channel
	mapCacheRegistry = s.zipTTL(mapCacheRegistry, <-ttlMapChan)

	return mapCacheRegistry, nil // err=nil by default, if everything is alright
}

//Recover current ttl information about registry
func (s RedisCacheStorage) GetTTL(key string) (int, error) {
	oneItemMap := make(map[string]CacheRegistry, 1)

	oneItemMap[key] = CacheRegistry{key, "", -2 /*not found*/, true}

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
func (s RedisCacheStorage) GetActualTTL(mapCacheRegistry map[string]CacheRegistry) (map[string]CacheRegistry, error) {

	conn := s.redisPool.Get()
	defer conn.Close()

	//prepare a keyval pair array
	for keyMap, cacheRegistry := range mapCacheRegistry {

		respTtl, err := conn.Do("ttl", s.getKey(keyMap))
		log.Debug("TTL %v that came from redis %v", keyMap, respTtl)

		if err != nil {
			log.Error("Error trying to retrieve ttl of key "+keyMap, err)
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
func (s RedisCacheStorage) GetTTLMap(keys []string) map[string]int {

	ttlMap := make(map[string]int, len(keys))

	conn := s.redisPool.Get()
	defer conn.Close()

	//prepare a keyval pair array
	for _, key := range keys {

		respTtl, err := conn.Do("ttl", s.getKey(key))
		log.Debug("TTL %v that came from redis %v", key, respTtl)

		if err != nil {
			log.Error("Error trying to retrieve ttl of key "+key, err)
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
func (s RedisCacheStorage) SetValues(registries ...CacheRegistry) error {

	var cacheRegistry CacheRegistry
	var index int

	defer func(cacheRegistry *CacheRegistry) {
		if r := recover(); r != nil {
			log.Error("Error trying to save cacheRegistry! recover= %v", r)
		}
	}(&cacheRegistry)

	conn := s.redisPool.Get()
	defer conn.Close()

	keyValPairs := make([]interface{}, 2*len(registries))

	//prepare a keyval pair array
	for index, cacheRegistry = range registries {
		buffer := new(bytes.Buffer)

		if len(cacheRegistry.CacheKey) == 0 {
			log.Error("cache key vazio !!!")
			//panic(errors.New("cache key vazio"))
		}

		buffer.Reset()
		e := gob.NewEncoder(buffer)
		err := e.Encode(cacheRegistry)
		if err != nil {
			log.Error("Error trying to save registry! %v", err)
			return err
		}

		bytes := buffer.Bytes()

		if len(bytes) == 0 {
			log.Error("Error trying to decode value for key %v", cacheRegistry.CacheKey)
		}

		keyValPairs[(index * 2)] = s.getKey(cacheRegistry.CacheKey)
		keyValPairs[(index*2)+1] = bytes

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
	conn := s.redisPool.Get()
	defer conn.Close()

	//prepare a keyval pair array
	for _, cacheRegistry := range cacheRegistries {
		if cacheRegistry.GetTTL() > 0 {
			log.Debug("Setting ttl to key %s ", cacheRegistry.CacheKey)
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
func (s RedisCacheStorage) DeleteValues(cacheKeys ...string) error {

	c := s.redisPool.Get()
	defer func() {
		c.Close()
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

	if len(s.cacheAreaa) > 0 {
		newKey = s.cacheAreaa + ":" + key
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
func NewRedisCacheStorage(hostPort string, password string, maxIdle int, cacheArea string) RedisCacheStorage {

	redisCacheStorage := RedisCacheStorage{
		*newPoolRedis(hostPort, password, maxIdle),
		cacheArea,
	}

	return redisCacheStorage
}

//create a redis connection pool
func newPoolRedis(server, password string, maxIdle int) *redis.Pool {

	return &redis.Pool{
		MaxIdle:     maxIdle,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {

			c, err := redis.Dial("tcp", server)
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

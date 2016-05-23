package cache

import (
	//"bytes"
	//"encoding/gob"
	"time"

	"github.com/garyburd/redigo/redis"
)

//Cache storage implementation using redis as key/value storage
type RedisCacheStorage struct {
	redisPool      redis.Pool
	ttlReadTimeout int
	cacheArea      string
	Serializer     Serializer // usually SerializerGOB implementation
}

var _ = SerializerGOB{} // this is the usual serializer used above!!


//recover all cacheregistries of keys
func (s RedisCacheStorage) GetValuesMap(cacheKeys ...string) (map[string]CacheRegistry, error) {

	mapCacheRegistry := make(map[string]CacheRegistry)

	if len(cacheKeys) <= 0 {
		log.Debug("Nenhuma chave informada para busca. len(arrKeys)=0!")
		return mapCacheRegistry, nil //empty map
	}

	conn := s.redisPool.Get()
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

			cacheRegistryBytes, errBytes := redis.Bytes(cacheRegistryNotBytes, err)
			if errBytes != nil || replyMget == nil {
				return mapCacheRegistry, errBytes
			}

			cacheRegistry := CacheRegistry{}

			interfaceResp, _, errUnm := s.Serializer.UnmarshalMsg(cacheRegistry, cacheRegistryBytes)
			if errUnm != nil {
				log.Error("error trying to deserialize!", errUnm)
				return mapCacheRegistry, errUnm
			}

			cacheRegistry, isCR := interfaceResp.(CacheRegistry)
			if (!isCR) {
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

	return mapCacheRegistry, nil // err=nil by default, if everything is alright
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

	keyValPairs := make([]interface{}, 2 * len(registries))

	//prepare a keyval pair array
	for index, cacheRegistry = range registries {

		if len(cacheRegistry.CacheKey) == 0 {
			log.Error("cache key vazio !!!")
			//panic(errors.New("cache key vazio"))
		}

		var bytes = []byte{}
		bytes, err := s.Serializer.MarshalMsg(cacheRegistry, bytes)
		if (err != nil) {
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
	log.Debug("Calling set ttl")
	s.SetExpireTTL(registries...)
	return nil
}

//set defined ttl to the cache registries
func (s RedisCacheStorage) SetExpireTTL(cacheRegistries ...CacheRegistry) {
	conn := s.redisPool.Get()
	defer conn.Close()

	//prepare a keyval pair array
	for _, cacheRegistry := range cacheRegistries {
		if cacheRegistry.StorageTTL > 0 {
			log.Debug("Setting ttl to key %s ", cacheRegistry.CacheKey)
			_, err := conn.Do("expire", s.getKey(cacheRegistry.CacheKey), cacheRegistry.StorageTTL)
			if err != nil {
				log.Error("Error trying to save cache registry w! %v", err)
				return
			}

		} else {
			log.Debug("TTL for %s, ttl=%d will not be setted! ", s.getKey(cacheRegistry.CacheKey), cacheRegistry.StorageTTL)
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

	var serPredix = s.Serializer.GetPrefix()

	if len(s.cacheArea) > 0 {
		newKey = s.cacheArea + ":" + serPredix + ":" + key
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
		serializer,
	}

	return redisCacheStorage
}

//create a redis connection pool
func newPoolRedis(server, password string, maxIdle int, readTimeout int) *redis.Pool {

	return &redis.Pool{
		MaxIdle:     maxIdle,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {

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

package aop

import (
	"encoding/gob"
	"fmt"
	"github.com/darciopacifico/enablecache/cache"
	"math/rand"
	"strconv"
	"testing"
	"sync"
	"time"
)

type UserEmail struct {
	Email string
	Name  string
	Uuid  string
}

func (u UserEmail) GetCacheKey() string {
	return "User:" + u.Email
}

var (
	csRedis = cache.NewRedisCacheStorage("localhost:6379", "", 8,200, 2000,  "str_test", cache.SerializerGOB{}, true)
	cmStr   = cache.SimpleCacheManager{
		CacheStorage: csRedis,
	}

	cs CacheSpot
)

var CFindByEmail func(email string) UserEmail

func init() {

	cs = CacheSpot{
		CachedFunc:   &CFindByEmail,
		HotFunc:      FindByEmail,
		Ttl: 100 * time.Second,
		CacheManager: cmStr,
		WaitingGroup: &sync.WaitGroup{},
	}.MustStartCache()

	gob.Register(UserEmail{})
}

func FindByEmail(email string) UserEmail {
	return UserEmail{
		Name:  "some name",
		Email: email,
		Uuid:  strconv.Itoa(rand.Int()),
	}
}

func TestFindByStr(t *testing.T) {
	defer cs.WaitAllParallelOps()

	user1 := CFindByEmail("darcio.paciico@gmail.com")
	fmt.Println(" %v", user1)

	user2 := CFindByEmail("darcio.paciico@gmail.com")
	fmt.Println(" %v", user2)

	user3 := CFindByEmail("darcio.paciico@gmail.com")

	fmt.Println(" %v", user3)
	/*
	 */

}

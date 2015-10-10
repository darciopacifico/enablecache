package aop

import (
	"gitlab.wmxp.com.br/bis/biro/cache"
	"gitlab.wmxp.com.br/bis/biro/config"

	"math/rand"
	"strconv"
	"time"
)

var (
	conf = config.CreateConfig()

	cacheStorage = cache.NewRedisCacheStorage(conf.Config("ipPortRedis", "localhost:6379"), conf.Config("passwordRedis", ""), 8, strconv.Itoa(rand.Int()))
	cacheManager = cache.SimpleCacheManager{
		Ps: cacheStorage,
	}

	cacheManagerUpdater = cache.UpdaterCacheManagerImpl{
		SimpleCacheManager: cacheManager,
	}
)

type User struct {
	Id       int
	Name     string
	Creation time.Time
	Ttl      int
}

func (u User) GetTtl() int {
	return u.Ttl
}
func (u User) SetTtl(ttl int) interface{} {
	u.Ttl = ttl
	return u
}

func FindUser(id int) User {

	user := createUser(id)
	user.Ttl = 1 // by some fake business rule, the ttl will be one second
	return user
}

func createUser(id int) User {
	user := User{
		id,
		"User:" + strconv.Itoa(id),
		time.Now(),
		-1,
	}

	return user
}

type Customer struct {
	Id       int
	Name     string
	Creation time.Time
	Policies []InsurantePolicy
}

type InsurantePolicy struct {
	Id          int
	Description string
	Creation    time.Time
	Items       []IPItem
}

type IPItem struct {
	Id       int
	Name     string
	Creation time.Time
}

func FindCustomer(id int, name string, isActive bool) (Customer, bool, time.Time) {
	customer := createCustomer(id)
	return customer, true, customer.Creation
}

func FindCustomerSimple(id int) Customer {
	return createCustomer(id)
}

func createCustomer(id int) Customer {
	name := "Cliente=" + strconv.Itoa(id)
	return Customer{
		id,
		name,
		time.Now(),
		[]InsurantePolicy{
			createPolicy(1),
			createPolicy(2),
			createPolicy(3),
		},
	}
}

func createPolicy(id int) InsurantePolicy {
	name := "Apolice=" + strconv.Itoa(id)
	return InsurantePolicy{
		id,
		name,
		time.Now(),
		[]IPItem{
			createItem(1),
			createItem(2),
			createItem(3),
		},
	}
}

func createItem(id int) IPItem {
	name := "ItemAp=" + strconv.Itoa(id)
	return IPItem{
		id,
		name,
		time.Now(),
	}
}

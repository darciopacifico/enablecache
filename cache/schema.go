package cache
import (
	"fmt"
	"time"
)

//go:generate msgp

type Attribute struct {
	Id    int
	Name  string
	Value string
}

type Car struct {
	CarId      	int          		`json:"carId"`
	CarName    	string       		`json:"carName"`
	Attributes 	[]Attribute  		`json:"specializations,omitempty"`
	FlagMap    	map[string]string 	`json:"flagMap"`
	Ttl        	int            		`json:"-"`

}


func GetReg(id int) CacheRegistry {
	cr := CacheRegistry{
		CacheKey: fmt.Sprintf("cacheReg_%v", id),
		Payload :GetCar(id),
		StoreTTL: 3600,
		CacheTime: time.Now(),
		HasValue:true,
		TypeName:"",
	}
	return cr
}


func GetCar(id int) Car {
	var cr =Car{
		CarId    : id,
		CarName  : "BMW 540",
		Attributes : []Attribute{
			Attribute{1, "deckoker", "true", },
			Attribute{2, "actionMode", "simple/double", },
			Attribute{3, "lock", "bolt locker", },
			Attribute{1, "deckoker", "true", },
			Attribute{2, "actionMode", "simple/double", },
			Attribute{2, "actionMode", "simple/double", },
			Attribute{2, "actionMode", "simple/double", },
			Attribute{3, "lock", "bolt locker", },
			Attribute{1, "deckoker", "true", },
			Attribute{2, "actionMode", "simple/double", },
			Attribute{2, "actionMode", "simple/double", },
			Attribute{3, "lock", "bolt locker", },
		},
		Ttl             : 700000,
	}
	return cr
}


func GetRegs(qtd int) []CacheRegistry {
	cr := make([]CacheRegistry, qtd)
	for i := 0; i < qtd; i++ {
		cr[i] = GetReg(i)
	}
	return cr
}

func GetKeys(qtd int) []string {
	keys := make([]string, qtd)
	for i := 0; i < qtd; i++ {
		keys[i] = fmt.Sprintf("cacheReg_%v", i)
	}
	return keys
}


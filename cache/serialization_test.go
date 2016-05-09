package cache



import (
	"testing"
	"encoding/gob"
	"fmt"
	"bytes"
	"encoding/json"
	"gopkg.in/vmihailenco/msgpack.v2"
)

func init() {
	gob.Register(Car{})
	gob.Register(Attribute{})

	cacheStorageRedis_test.SetValues(GetRegs(qtdChaves)...)

}



func TestMGET(t *testing.T) {

	cacheReg, err := cacheStorageRedis_test.GetValuesMap(GetKeys(qtdChaves)...)

	if (err != nil || len(cacheReg) == 0) {
		log.Error("Error values not found")
		t.Fail()
	}
}


func TestMultiMsgPack(t *testing.T) {

	var gun = GetReg(1)

	bytes,errM := msgpack.Marshal(gun)
	if(errM!=nil){
		fmt.Println(errM)
	}

	var gunUnm = CacheRegistry{Payload:&Car{}}

	errU := msgpack.Unmarshal(bytes,&gunUnm)
	if(errU!=nil){
		fmt.Println(errU)
	}

	gunDecoded,_ := gunUnm.Payload.(*Car)

	fmt.Println("gunName v%",gunDecoded.CarName)
}

func BenchmarkMget(b *testing.B) {

	var cacheReg map[string]CacheRegistry
	var err error

	for i := 0; i < b.N; i++ {
		cacheReg, err = cacheStorageRedis_test.GetValuesMap(GetKeys(qtdChaves)...)

		if (err != nil || len(cacheReg) == 0) {
			log.Error("Error values not found")
			b.Fail()
		}
	}

	for k, _ := range cacheReg {

		if (cacheReg[k].Payload == nil) {
			b.Fail()
		}else {
			//fmt.Println(cacheReg[k].CacheKey)
		}

	}
}



func TestMsgp(t *testing.T){

	var carOriginal = GetCar(1)

	bytes := make([]byte,0)

	bytesFromCar, err := carOriginal.MarshalMsg(bytes)
	if(err!=nil){
		fmt.Println("Erro %v",err)
	}

	var carDecoded = Car{}

	carDecoded.UnmarshalMsg(bytesFromCar)

	fmt.Println("id   		= ", carDecoded.CarId)
	fmt.Println("name 		= ", carDecoded.CarName)
	fmt.Println("len att 	= ", len(carDecoded.Attributes))

}


func BenchmarkCarMSGP(b *testing.B) {
	var cacheOriginal = GetCar(987987)

	bytes := make([]byte,0)
	bytesFromCar, err := cacheOriginal.MarshalMsg(bytes)
	if(err!=nil){
		fmt.Println("Erro %v",err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var carDecoded = Car{}
		carDecoded.UnmarshalMsg(bytesFromCar)
		if(carDecoded.CarId<=0 || len(carDecoded.CarName)<1 ){
			b.Fail()
			fmt.Println("id   = ", carDecoded.CarId)
			fmt.Println("name = ", carDecoded.CarName)
		}
	}
}


func BenchmarkCarGOB(b *testing.B) {
	var carOriginal = GetCar(234234234)

	var destBytes []byte
	bufferE := bytes.NewBuffer(destBytes)
	e := gob.NewEncoder(bufferE)
	e.Encode(carOriginal)
	destBytes = bufferE.Bytes()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		var carDecoded = Car{}

		reader := bytes.NewReader(destBytes)

		d:=gob.NewDecoder(reader)

		d.Decode(&carDecoded)

		if(carDecoded.CarId<=0 || len(carDecoded.CarName)<1 ){
			b.Fail()
			fmt.Println("id   = ", carDecoded.CarId)
			fmt.Println("name = ", carDecoded.CarName)
		}
	}
}


func BenchmarkCarJSON(b *testing.B) {
	var carOriginal = GetCar(234234234)



	var destBytes []byte
	bufferE := bytes.NewBuffer(destBytes)
	e := json.NewEncoder(bufferE)

	e.Encode(carOriginal)
	destBytes = bufferE.Bytes()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		var carDecoded = Car{}

		reader := bytes.NewReader(destBytes)

		d:=json.NewDecoder(reader)

		d.Decode(&carDecoded)

		if(carDecoded.CarId<=0 || len(carDecoded.CarName)<1 ){
			b.Fail()
			fmt.Println("id   = ", carDecoded.CarId)
			fmt.Println("name = ", carDecoded.CarName)
		}
	}
}


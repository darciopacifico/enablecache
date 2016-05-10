package cache
import (
	"bytes"
	"encoding/gob"
	"reflect"
)


type SerializerGOB struct {
	MapSerializer map[string]reflect.Type
}

//
func (s SerializerGOB) Register(value interface{}){
	gob.Register(value)
	name, rt := getNameType(value)
	s.MapSerializer[name]=rt
}

//
func getNameType(value interface{})(string, reflect.Type){
	rt := reflect.TypeOf(value)
	name := rt.String()

	star := ""
	if rt.Name() == "" {
		if pt := rt; pt.Kind() == reflect.Ptr {
			star = "*"
			rt = pt
		}
	}
	if rt.Name() != "" {
		if rt.PkgPath() == "" {
			name = star + rt.Name()
		} else {
			name = star + rt.PkgPath() + "." + rt.Name()
		}
	}

	return name, rt
}


// seralize an objeto to byte array
func (SerializerGOB) MarshalMsg(src CacheRegistry, b []byte) (o []byte, err error){

	typeName, _ := getNameType(src.Payload)

	src.TypeName = typeName

	buffer := new(bytes.Buffer)
	buffer.Reset()
	e := gob.NewEncoder(buffer)
	err = e.Encode(src)
	if err != nil {
		log.Error("Error trying to save registry! %v", err)
		return []byte{}, err
	}
	bytes := buffer.Bytes()
	return bytes,err
}

// deserialize an byte array to object
func (SerializerGOB) UnmarshalMsg(dest CacheRegistry, bts []byte) (resp interface{}, o []byte, err error){
	bufferResp := bytes.NewBuffer(bts)

	d := gob.NewDecoder(bufferResp) //instantiate a decoder base on bytes
	err = d.Decode(&dest) // try to decode this bytes in a cacheRegistry object

	return dest, bts, err
}

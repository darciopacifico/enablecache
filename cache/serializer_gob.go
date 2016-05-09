package cache
import (
	"bytes"
	"encoding/gob"
)


type SerializerGOB struct {

}

// seralize an objeto to byte array
func (SerializerGOB) MarshalMsg(src CacheRegistry, b []byte) (o []byte, err error){

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



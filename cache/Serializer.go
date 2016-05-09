package cache


type Serializer  interface {

	//register the type for future reference
	Register(sample interface{})

	// MarshalMsg implements msgp.Marshaler
	MarshalMsg(src CacheRegistry, b []byte) (o []byte, err error)

	// UnmarshalMsg implements msgp.Unmarshaler
	UnmarshalMsg(dest CacheRegistry, bts []byte) (resp interface{} ,o []byte, err error)


}
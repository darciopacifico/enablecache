package cache


type Serializer  interface {


	// MarshalMsg implements msgp.Marshaler
	MarshalMsg(src interface{}, b []byte) (o []byte, err error)

	// UnmarshalMsg implements msgp.Unmarshaler
	UnmarshalMsg(dest interface{}, bts []byte) (resp interface{} ,o []byte, err error)


}
package cache

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import "github.com/tinylib/msgp/msgp"

// DecodeMsg implements msgp.Decodable
func (z *CacheRegistry) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var xvk uint32
	xvk, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for xvk > 0 {
		xvk--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "CacheKey":
			z.CacheKey, err = dc.ReadString()
			if err != nil {
				return
			}
		case "Payload":
			z.Payload, err = dc.ReadIntf()
			if err != nil {
				return
			}
		case "StorageTTL":
			z.StorageTTL, err = dc.ReadFloat64()
			if err != nil {
				return
			}
		case "CacheTime":
			z.CacheTime, err = dc.ReadTime()
			if err != nil {
				return
			}
		case "HasValue":
			z.HasValue, err = dc.ReadBool()
			if err != nil {
				return
			}
		case "TypeName":
			z.TypeName, err = dc.ReadString()
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *CacheRegistry) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 6
	// write "CacheKey"
	err = en.Append(0x86, 0xa8, 0x43, 0x61, 0x63, 0x68, 0x65, 0x4b, 0x65, 0x79)
	if err != nil {
		return err
	}
	err = en.WriteString(z.CacheKey)
	if err != nil {
		return
	}
	// write "Payload"
	err = en.Append(0xa7, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteIntf(z.Payload)
	if err != nil {
		return
	}
	// write "StorageTTL"
	err = en.Append(0xaa, 0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x54, 0x54, 0x4c)
	if err != nil {
		return err
	}
	err = en.WriteFloat64(z.StorageTTL)
	if err != nil {
		return
	}
	// write "CacheTime"
	err = en.Append(0xa9, 0x43, 0x61, 0x63, 0x68, 0x65, 0x54, 0x69, 0x6d, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteTime(z.CacheTime)
	if err != nil {
		return
	}
	// write "HasValue"
	err = en.Append(0xa8, 0x48, 0x61, 0x73, 0x56, 0x61, 0x6c, 0x75, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.HasValue)
	if err != nil {
		return
	}
	// write "TypeName"
	err = en.Append(0xa8, 0x54, 0x79, 0x70, 0x65, 0x4e, 0x61, 0x6d, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(z.TypeName)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *CacheRegistry) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 6
	// string "CacheKey"
	o = append(o, 0x86, 0xa8, 0x43, 0x61, 0x63, 0x68, 0x65, 0x4b, 0x65, 0x79)
	o = msgp.AppendString(o, z.CacheKey)
	// string "Payload"
	o = append(o, 0xa7, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64)
	o, err = msgp.AppendIntf(o, z.Payload)
	if err != nil {
		return
	}
	// string "StorageTTL"
	o = append(o, 0xaa, 0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x54, 0x54, 0x4c)
	o = msgp.AppendFloat64(o, z.StorageTTL)
	// string "CacheTime"
	o = append(o, 0xa9, 0x43, 0x61, 0x63, 0x68, 0x65, 0x54, 0x69, 0x6d, 0x65)
	o = msgp.AppendTime(o, z.CacheTime)
	// string "HasValue"
	o = append(o, 0xa8, 0x48, 0x61, 0x73, 0x56, 0x61, 0x6c, 0x75, 0x65)
	o = msgp.AppendBool(o, z.HasValue)
	// string "TypeName"
	o = append(o, 0xa8, 0x54, 0x79, 0x70, 0x65, 0x4e, 0x61, 0x6d, 0x65)
	o = msgp.AppendString(o, z.TypeName)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *CacheRegistry) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var bzg uint32
	bzg, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for bzg > 0 {
		bzg--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "CacheKey":
			z.CacheKey, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "Payload":
			z.Payload, bts, err = msgp.ReadIntfBytes(bts)
			if err != nil {
				return
			}
		case "StorageTTL":
			z.StorageTTL, bts, err = msgp.ReadFloat64Bytes(bts)
			if err != nil {
				return
			}
		case "CacheTime":
			z.CacheTime, bts, err = msgp.ReadTimeBytes(bts)
			if err != nil {
				return
			}
		case "HasValue":
			z.HasValue, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				return
			}
		case "TypeName":
			z.TypeName, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z *CacheRegistry) Msgsize() (s int) {
	s = 1 + 9 + msgp.StringPrefixSize + len(z.CacheKey) + 8 + msgp.GuessSize(z.Payload) + 11 + msgp.Float64Size + 10 + msgp.TimeSize + 9 + msgp.BoolSize + 9 + msgp.StringPrefixSize + len(z.TypeName)
	return
}

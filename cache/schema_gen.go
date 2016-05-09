package cache

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import "github.com/tinylib/msgp/msgp"

// DecodeMsg implements msgp.Decodable
func (z *Attribute) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "Id":
			z.Id, err = dc.ReadInt()
			if err != nil {
				return
			}
		case "Name":
			z.Name, err = dc.ReadString()
			if err != nil {
				return
			}
		case "Value":
			z.Value, err = dc.ReadString()
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
func (z Attribute) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "Id"
	err = en.Append(0x83, 0xa2, 0x49, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.Id)
	if err != nil {
		return
	}
	// write "Name"
	err = en.Append(0xa4, 0x4e, 0x61, 0x6d, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Name)
	if err != nil {
		return
	}
	// write "Value"
	err = en.Append(0xa5, 0x56, 0x61, 0x6c, 0x75, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Value)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z Attribute) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "Id"
	o = append(o, 0x83, 0xa2, 0x49, 0x64)
	o = msgp.AppendInt(o, z.Id)
	// string "Name"
	o = append(o, 0xa4, 0x4e, 0x61, 0x6d, 0x65)
	o = msgp.AppendString(o, z.Name)
	// string "Value"
	o = append(o, 0xa5, 0x56, 0x61, 0x6c, 0x75, 0x65)
	o = msgp.AppendString(o, z.Value)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Attribute) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "Id":
			z.Id, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				return
			}
		case "Name":
			z.Name, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "Value":
			z.Value, bts, err = msgp.ReadStringBytes(bts)
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

func (z Attribute) Msgsize() (s int) {
	s = 1 + 3 + msgp.IntSize + 5 + msgp.StringPrefixSize + len(z.Name) + 6 + msgp.StringPrefixSize + len(z.Value)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *Car) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var wht uint32
	wht, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for wht > 0 {
		wht--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "CarId":
			z.CarId, err = dc.ReadInt()
			if err != nil {
				return
			}
		case "CarName":
			z.CarName, err = dc.ReadString()
			if err != nil {
				return
			}
		case "Attributes":
			var hct uint32
			hct, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.Attributes) >= int(hct) {
				z.Attributes = z.Attributes[:hct]
			} else {
				z.Attributes = make([]Attribute, hct)
			}
			for bai := range z.Attributes {
				var cua uint32
				cua, err = dc.ReadMapHeader()
				if err != nil {
					return
				}
				for cua > 0 {
					cua--
					field, err = dc.ReadMapKeyPtr()
					if err != nil {
						return
					}
					switch msgp.UnsafeString(field) {
					case "Id":
						z.Attributes[bai].Id, err = dc.ReadInt()
						if err != nil {
							return
						}
					case "Name":
						z.Attributes[bai].Name, err = dc.ReadString()
						if err != nil {
							return
						}
					case "Value":
						z.Attributes[bai].Value, err = dc.ReadString()
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
			}
		case "FlagMap":
			var xhx uint32
			xhx, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.FlagMap == nil && xhx > 0 {
				z.FlagMap = make(map[string]string, xhx)
			} else if len(z.FlagMap) > 0 {
				for key, _ := range z.FlagMap {
					delete(z.FlagMap, key)
				}
			}
			for xhx > 0 {
				xhx--
				var cmr string
				var ajw string
				cmr, err = dc.ReadString()
				if err != nil {
					return
				}
				ajw, err = dc.ReadString()
				if err != nil {
					return
				}
				z.FlagMap[cmr] = ajw
			}
		case "Ttl":
			z.Ttl, err = dc.ReadInt()
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
func (z *Car) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 5
	// write "CarId"
	err = en.Append(0x85, 0xa5, 0x43, 0x61, 0x72, 0x49, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.CarId)
	if err != nil {
		return
	}
	// write "CarName"
	err = en.Append(0xa7, 0x43, 0x61, 0x72, 0x4e, 0x61, 0x6d, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(z.CarName)
	if err != nil {
		return
	}
	// write "Attributes"
	err = en.Append(0xaa, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.Attributes)))
	if err != nil {
		return
	}
	for bai := range z.Attributes {
		// map header, size 3
		// write "Id"
		err = en.Append(0x83, 0xa2, 0x49, 0x64)
		if err != nil {
			return err
		}
		err = en.WriteInt(z.Attributes[bai].Id)
		if err != nil {
			return
		}
		// write "Name"
		err = en.Append(0xa4, 0x4e, 0x61, 0x6d, 0x65)
		if err != nil {
			return err
		}
		err = en.WriteString(z.Attributes[bai].Name)
		if err != nil {
			return
		}
		// write "Value"
		err = en.Append(0xa5, 0x56, 0x61, 0x6c, 0x75, 0x65)
		if err != nil {
			return err
		}
		err = en.WriteString(z.Attributes[bai].Value)
		if err != nil {
			return
		}
	}
	// write "FlagMap"
	err = en.Append(0xa7, 0x46, 0x6c, 0x61, 0x67, 0x4d, 0x61, 0x70)
	if err != nil {
		return err
	}
	err = en.WriteMapHeader(uint32(len(z.FlagMap)))
	if err != nil {
		return
	}
	for cmr, ajw := range z.FlagMap {
		err = en.WriteString(cmr)
		if err != nil {
			return
		}
		err = en.WriteString(ajw)
		if err != nil {
			return
		}
	}
	// write "Ttl"
	err = en.Append(0xa3, 0x54, 0x74, 0x6c)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.Ttl)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Car) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 5
	// string "CarId"
	o = append(o, 0x85, 0xa5, 0x43, 0x61, 0x72, 0x49, 0x64)
	o = msgp.AppendInt(o, z.CarId)
	// string "CarName"
	o = append(o, 0xa7, 0x43, 0x61, 0x72, 0x4e, 0x61, 0x6d, 0x65)
	o = msgp.AppendString(o, z.CarName)
	// string "Attributes"
	o = append(o, 0xaa, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Attributes)))
	for bai := range z.Attributes {
		// map header, size 3
		// string "Id"
		o = append(o, 0x83, 0xa2, 0x49, 0x64)
		o = msgp.AppendInt(o, z.Attributes[bai].Id)
		// string "Name"
		o = append(o, 0xa4, 0x4e, 0x61, 0x6d, 0x65)
		o = msgp.AppendString(o, z.Attributes[bai].Name)
		// string "Value"
		o = append(o, 0xa5, 0x56, 0x61, 0x6c, 0x75, 0x65)
		o = msgp.AppendString(o, z.Attributes[bai].Value)
	}
	// string "FlagMap"
	o = append(o, 0xa7, 0x46, 0x6c, 0x61, 0x67, 0x4d, 0x61, 0x70)
	o = msgp.AppendMapHeader(o, uint32(len(z.FlagMap)))
	for cmr, ajw := range z.FlagMap {
		o = msgp.AppendString(o, cmr)
		o = msgp.AppendString(o, ajw)
	}
	// string "Ttl"
	o = append(o, 0xa3, 0x54, 0x74, 0x6c)
	o = msgp.AppendInt(o, z.Ttl)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Car) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var lqf uint32
	lqf, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for lqf > 0 {
		lqf--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "CarId":
			z.CarId, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				return
			}
		case "CarName":
			z.CarName, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "Attributes":
			var daf uint32
			daf, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Attributes) >= int(daf) {
				z.Attributes = z.Attributes[:daf]
			} else {
				z.Attributes = make([]Attribute, daf)
			}
			for bai := range z.Attributes {
				var pks uint32
				pks, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				for pks > 0 {
					pks--
					field, bts, err = msgp.ReadMapKeyZC(bts)
					if err != nil {
						return
					}
					switch msgp.UnsafeString(field) {
					case "Id":
						z.Attributes[bai].Id, bts, err = msgp.ReadIntBytes(bts)
						if err != nil {
							return
						}
					case "Name":
						z.Attributes[bai].Name, bts, err = msgp.ReadStringBytes(bts)
						if err != nil {
							return
						}
					case "Value":
						z.Attributes[bai].Value, bts, err = msgp.ReadStringBytes(bts)
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
			}
		case "FlagMap":
			var jfb uint32
			jfb, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.FlagMap == nil && jfb > 0 {
				z.FlagMap = make(map[string]string, jfb)
			} else if len(z.FlagMap) > 0 {
				for key, _ := range z.FlagMap {
					delete(z.FlagMap, key)
				}
			}
			for jfb > 0 {
				var cmr string
				var ajw string
				jfb--
				cmr, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				ajw, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				z.FlagMap[cmr] = ajw
			}
		case "Ttl":
			z.Ttl, bts, err = msgp.ReadIntBytes(bts)
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

func (z *Car) Msgsize() (s int) {
	s = 1 + 6 + msgp.IntSize + 8 + msgp.StringPrefixSize + len(z.CarName) + 11 + msgp.ArrayHeaderSize
	for bai := range z.Attributes {
		s += 1 + 3 + msgp.IntSize + 5 + msgp.StringPrefixSize + len(z.Attributes[bai].Name) + 6 + msgp.StringPrefixSize + len(z.Attributes[bai].Value)
	}
	s += 8 + msgp.MapHeaderSize
	if z.FlagMap != nil {
		for cmr, ajw := range z.FlagMap {
			_ = ajw
			s += msgp.StringPrefixSize + len(cmr) + msgp.StringPrefixSize + len(ajw)
		}
	}
	s += 4 + msgp.IntSize
	return
}

package resp

import (
	"fmt"
	"strconv"
)

func (v Value) Marshal() []byte {
	switch v.Typ {
	case "string":
		return v.marshalString()
	case "bulk":
		return v.marshalBulk()
	case "array":
		return v.marshalArray()
	case "error":
		return v.marshalError()
	case "null":
		return v.marshalNull()
	default:
		fmt.Printf("Unknown type: %v", v.Typ)
		return []byte{}
	}
}

func (v Value) marshalString() []byte {
	bytes := make([]byte, 0)
	bytes = append(bytes, '+')
	bytes = append(bytes, v.Str...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (v Value) marshalBulk() []byte {
	bytes := make([]byte, 0)
	len := len(v.Bulk)
	bytes = append(bytes, '$')
	bytes = append(bytes, strconv.Itoa(len)...)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, v.Bulk...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (v Value) marshalArray() []byte {
	bytes := make([]byte, 0)
	len := len(v.Array)
	bytes = append(bytes, '*')
	bytes = append(bytes, strconv.Itoa(len)...)
	bytes = append(bytes, '\r', '\n')

	for _, val := range v.Array {
		bytes = append(bytes, val.Marshal()...)
	}

	return bytes
}

func (v Value) marshalError() []byte {
	bytes := make([]byte, 0)
	bytes = append(bytes, ERROR)
	bytes = append(bytes, v.Str...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (v Value) marshalNull() []byte {
	return []byte("$-1\r\n")
}

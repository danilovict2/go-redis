package resp

import (
	"fmt"
	"strconv"
)

func (v Value) Marshal() []byte {
	switch v.Typ {
	case STRING_TYPE:
		return v.marshalString()
	case BULK_TYPE:
		return v.marshalBulk()
	case ARRAY_TYPE:
		return v.marshalArray()
	case ERROR_TYPE:
		return v.marshalError()
	case NULL_TYPE:
		return v.marshalNull()
	case INTEGER_TYPE:
		return v.marshalInt()
	case NULL_ARRAY:
		return v.marshalNullArray()
	default:
		fmt.Printf("Unknown type: %v", v.Typ)
		return []byte{}
	}
}

func (v Value) marshalString() []byte {
	bytes := make([]byte, 0)
	bytes = append(bytes, STRING)
	bytes = append(bytes, v.Str...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (v Value) marshalBulk() []byte {
	bytes := make([]byte, 0)
	len := len(v.Bulk)
	bytes = append(bytes, BULK)
	bytes = append(bytes, strconv.Itoa(len)...)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, v.Bulk...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (v Value) marshalArray() []byte {
	bytes := make([]byte, 0)
	len := len(v.Array)
	bytes = append(bytes, ARRAY)
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

func (v Value) marshalNullArray() []byte {
	return []byte("*-1\r\n")
}

func (v Value) marshalInt() []byte {
	bytes := make([]byte, 0)
	bytes = append(bytes, INTEGER)
	bytes = append(bytes, fmt.Append(nil, v.Int)...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

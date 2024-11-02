package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

const (
	BULK = '$'
	ARRAY = '*'
	STRING = '+'
	ERROR = '-'
)

type Resp struct {
	reader *bufio.Reader
}

type Value struct {
	typ string
	str string
	bulk string
	array []Value
}

func NewResp(rd io.Reader) *Resp {
	return &Resp{
		reader: bufio.NewReader(rd),
	}
}

func (r *Resp) readLine() (line []byte, n int, err error) {
	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, 0, err
		}
		
		n += 1
		line = append(line, b)
		if len(line) >= 2 && line[len(line) - 2] == '\r' {
			break
		}
	}
	
	return line[:len(line) - 2], n, nil
}

func (r *Resp) readInteger() (int, int, error) {
	line, n, err := r.readLine()
	if err != nil {
		return 0, 0, err
	}

	i64, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, n, err
	}

	return int(i64), n, nil
}


func (r *Resp) Read() (Value, error) {
	_type, err := r.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch _type {
	case BULK:
		return r.readBulk()
	case ARRAY:
		return r.readArray()
	default:
		fmt.Printf("Unknown type: %v", string(_type))
		return Value{}, nil
	}
}

func (r *Resp) readBulk() (Value, error) {
	v := Value{}
	v.typ = "bulk"
	_, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	bulk, _, err := r.readLine()
	if err != nil {
		return v, err
	}

	v.bulk = string(bulk)
	return v, nil
}

func (r *Resp) readArray() (Value, error) {
	val := Value{}
	val.typ = "array"

	len, _, err := r.readInteger()
	if err != nil {
		return val, err
	}

	val.array = make([]Value, 0)
	for i := 0; i < len; i++ {
		v, err := r.Read()
		if err != nil {
			return val, err
		}

		val.array = append(val.array, v)
	}

	return val, nil
}

func (v Value) Marshal() []byte {
	switch v.typ {
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
		fmt.Printf("Unknown type: %v", v.typ)
		return []byte{}
	}
}

func (v Value) marshalString() []byte {
	bytes := make([]byte, 0)
	bytes = append(bytes, '+')
	bytes = append(bytes, v.str...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (v Value) marshalBulk() []byte {
	bytes := make([]byte, 0)
	len := len(v.bulk)
	bytes = append(bytes, '$')
	bytes = append(bytes, strconv.Itoa(len)...)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, v.bulk...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (v Value) marshalArray() []byte {
	bytes := make([]byte, 0)
	len := len(v.array)
	bytes = append(bytes, '*')
	bytes = append(bytes, strconv.Itoa(len)...)
	bytes = append(bytes, '\r', '\n')

	for _, val := range v.array {
		bytes = append(bytes, val.Marshal()...)
	}

	return bytes
}

func (v Value) marshalError() []byte {
	bytes := make([]byte, 0)
	bytes = append(bytes, ERROR)
	bytes = append(bytes, v.str...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (v Value) marshalNull() []byte {
	return []byte("_\r\n")
}

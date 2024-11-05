package resp

import (
	"bufio"
	"fmt"
	"strconv"
)

type Resp struct {
	Reader *bufio.Reader
}

func (r *Resp) readLine() (line []byte, n int, err error) {
	for {
		b, err := r.Reader.ReadByte()
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
	_type, err := r.Reader.ReadByte()
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
	v.Typ = "bulk"
	_, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	bulk, _, err := r.readLine()
	if err != nil {
		return v, err
	}

	v.Bulk = string(bulk)
	return v, nil
}

func (r *Resp) readArray() (Value, error) {
	val := Value{}
	val.Typ = "array"

	len, _, err := r.readInteger()
	if err != nil {
		return val, err
	}

	val.Array = make([]Value, 0)
	for i := 0; i < len; i++ {
		v, err := r.Read()
		if err != nil {
			return val, err
		}

		val.Array = append(val.Array, v)
	}

	return val, nil
}
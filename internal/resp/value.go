package resp

type Value struct {
	Typ   string
	Str   string
	Bulk  string
	Array []Value
}

const (
	BULK = '$'
	ARRAY = '*'
	STRING = '+'
	ERROR = '-'
)
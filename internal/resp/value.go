package resp

type Value struct {
	Typ   string
	Str   string
	Bulk  string
	Array []Value
	Int   int
}

const (
	BULK    = '$'
	ARRAY   = '*'
	STRING  = '+'
	ERROR   = '-'
	INTEGER = ':'
)

const (
	BULK_TYPE    = "bulk"
	ARRAY_TYPE   = "array"
	STRING_TYPE  = "string"
	ERROR_TYPE   = "error"
	NULL_TYPE    = "null"
	INTEGER_TYPE = "int"
)

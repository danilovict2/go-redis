package main

var Handlers = map[string]func([]Value) Value{
	"PING": ping,
	"ECHO": echo,
}

func ping(args []Value) Value {
	return Value{typ: "string", str: "PONG"}
}

func echo(args []Value) Value {
	if len(args) != 1 {
		return Value{typ: "error", str: "ERR wrong number of args for 'echo' command"}
	}

	return Value{typ: "string", str: args[0].bulk}
}

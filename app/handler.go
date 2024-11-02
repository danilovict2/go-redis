package main

import "sync"

var Handlers = map[string]func([]Value) Value{
	"PING": ping,
	"ECHO": echo,
	"SET": set,
	"GET": get,
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

var SETs = map[string]string{}
var SETsMu = sync.RWMutex{}

func set(args []Value) Value {
	if len(args) != 2 {
		return Value{typ: "error", str: "ERR wrong number of args for 'set' command"}
	}

	key := args[0].bulk
	value := args[1].bulk
	SETsMu.Lock()
	SETs[key] = value
	SETsMu.Unlock()

	return Value{typ: "string", str: "OK"}
}

func get(args []Value) Value {
	if len(args) != 1 {
		return Value{typ: "error", str: "ERR wrong number of args for 'set' command"}
	}

	key := args[0].bulk
	SETsMu.RLock()
	val, ok := SETs[key]
	SETsMu.RUnlock()

	if !ok {
		return Value{typ: "null"}
	}
	return Value{typ: "bulk", bulk: val}
}


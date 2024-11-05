package main

import (
	"strconv"
	"sync"
	"time"
)

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
	if len(args) != 2 && len(args) != 4 {
		return Value{typ: "error", str: "ERR wrong number of args for 'set' command"}
	}

	key := args[0].bulk
	value := args[1].bulk
	SETsMu.Lock()
	defer SETsMu.Unlock()

	SETs[key] = value
	if len(args) == 4 {
		if args[2].bulk != "px" {
			return Value{typ: "error", str: "ERR syntax error"}
		}

		i64, err := strconv.ParseInt(args[3].bulk, 10, 64)
		if err != nil {
			return Value{typ: "error", str: "value is not an integer or out of range"}
		}

		go func () {
			time.Sleep(time.Millisecond * time.Duration(i64))
			unset(key)
		}()
	}

	return Value{typ: "string", str: "OK"}
}

func unset(key string) {
	SETsMu.Lock()
	delete(SETs, key)
	SETsMu.Unlock()
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


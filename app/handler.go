package main

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"strconv"
	"sync"
	"time"
)

var Handlers = map[string]func([]resp.Value) resp.Value{
	"PING":   ping,
	"ECHO":   echo,
	"SET":    set,
	"GET":    get,
	"CONFIG": config,
	"KEYS":   keys,
}

func ping(args []resp.Value) resp.Value {
	return resp.Value{Typ: "string", Str: "PONG"}
}

func echo(args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.Value{Typ: "error", Str: "ERR wrong number of args for 'echo' command"}
	}

	return resp.Value{Typ: "string", Str: args[0].Bulk}
}

var SETs = map[string]string{}
var SETsMu = sync.RWMutex{}

func set(args []resp.Value) resp.Value {
	if len(args) != 2 && len(args) != 4 {
		return resp.Value{Typ: "error", Str: "ERR wrong number of args for 'set' command"}
	}

	key := args[0].Bulk
	value := args[1].Bulk
	SETsMu.Lock()
	defer SETsMu.Unlock()

	SETs[key] = value
	if len(args) == 4 {
		if args[2].Bulk != "px" {
			return resp.Value{Typ: "error", Str: "ERR syntax error"}
		}

		i64, err := strconv.ParseInt(args[3].Bulk, 10, 64)
		if err != nil {
			return resp.Value{Typ: "error", Str: "resp.value is not an integer or out of range"}
		}

		go func() {
			time.Sleep(time.Millisecond * time.Duration(i64))
			unset(key)
		}()
	}

	return resp.Value{Typ: "string", Str: "OK"}
}

func unset(key string) {
	SETsMu.Lock()
	delete(SETs, key)
	SETsMu.Unlock()
}

func get(args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.Value{Typ: "error", Str: "ERR wrong number of args for 'get' command"}
	}

	key := args[0].Bulk
	SETsMu.RLock()
	val, ok := SETs[key]
	SETsMu.RUnlock()

	if !ok {
		return resp.Value{Typ: "null"}
	}
	return resp.Value{Typ: "bulk", Bulk: val}
}

var CONFIGs = map[string]string{}
var CONFIGsMu = sync.RWMutex{}

func config(args []resp.Value) resp.Value {
	if len(args) != 2 {
		return resp.Value{Typ: "error", Str: "ERR wrong number of arguments for 'config' command"}
	}

	switch args[0].Bulk {
	case "GET":
		return configGet(args[1:])
	default:
		return resp.Value{Typ: "error", Str: fmt.Sprintf("ERR unknown subcommand '%v'", args[0].Bulk)}
	}
}

func configGet(args []resp.Value) resp.Value {
	key := args[0].Bulk
	CONFIGsMu.RLock()
	value, ok := CONFIGs[key]
	CONFIGsMu.RUnlock()

	ret := resp.Value{Typ: "array"}
	if ok {
		ret.Array = append(ret.Array, resp.Value{Typ: "bulk", Bulk: key}, resp.Value{Typ: "bulk", Bulk: value})
	}

	return ret
}

func keys(args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.Value{Typ: "error", Str: "ERR wrong number of arguments for 'keys' command"}
	}

	CONFIGsMu.RLock()
	path := CONFIGs["dir"] + "/" + CONFIGs["dbfilename"]
	CONFIGsMu.RUnlock()

	str, err := readFile(path)
	if err != nil {
		return resp.Value{Typ: "error", Str: err.Error()}
	}

	ret := resp.Value{Typ: "array"}
	if len(str) > 0 {
		ret.Array = append(ret.Array, resp.Value{Typ: "bulk", Bulk: str})
	}
	return ret
}

package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

var Handlers = map[string]func([]resp.Value) resp.Value{
	"PING":     ping,
	"ECHO":     echo,
	"SET":      set,
	"GET":      get,
	"CONFIG":   config,
	"KEYS":     keys,
	"INFO":     info,
	"REPLCONF": replconf,
	"PSYNC":    psync,
	"TYPE":     typ,
	"XADD":     xadd,
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
		var unit time.Duration
		switch strings.ToUpper(args[2].Bulk) {
		case "EX":
			unit = time.Second
		case "PX":
			unit = time.Millisecond
		default:
			return resp.Value{Typ: "error", Str: "ERR syntax error"}
		}

		i64, err := strconv.ParseInt(args[3].Bulk, 10, 64)
		if err != nil {
			return resp.Value{Typ: "error", Str: "value is not an integer or out of range"}
		}

		go func() {
			time.Sleep(unit * time.Duration(i64))
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
	value, ok := server.configs[key]

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

	ret := resp.Value{Typ: "array"}
	SETsMu.RLock()
	for key := range SETs {
		ret.Array = append(ret.Array, resp.Value{Typ: "bulk", Bulk: key})
	}
	SETsMu.RUnlock()

	return ret
}

func info(args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.Value{Typ: "error", Str: "ERR wrong number of arguments for 'info' command"}
	}

	switch strings.ToUpper(args[0].Bulk) {
	case "REPLICATION":
		return infoReplication()
	default:
		return resp.Value{Typ: "bulk"}
	}
}

func infoReplication() resp.Value {
	ret := resp.Value{Typ: "bulk"}
	replicaof := server.replconf.host

	if replicaof != "" {
		ret.Bulk = "role:slave\r\n"
	} else {
		ret.Bulk = "role:master\r\nmaster_replid:8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb\r\nmaster_repl_offset:0\r\n"
	}

	return ret
}

func replconf(args []resp.Value) resp.Value {
	switch strings.ToUpper(args[0].Bulk) {
	case "GETACK":
		return replconfgetack(args[1:])
	case "ACK":
		return resp.Value{}
	default:
		return resp.Value{Typ: "string", Str: "OK"}
	}
}

func replconfgetack(args []resp.Value) resp.Value {
	if args[0].Bulk != "*" {
		return resp.Value{Typ: "error", Str: "Invalid GETACK parameter"}
	}

	offset := strconv.Itoa(server.offset)
	ret := resp.Value{Typ: "array"}
	ret.Array = append(ret.Array, resp.Value{Typ: "bulk", Bulk: "REPLCONF"}, resp.Value{Typ: "bulk", Bulk: "ACK"}, resp.Value{Typ: "bulk", Bulk: offset})
	return ret
}

func psync(args []resp.Value) resp.Value {
	return resp.Value{Typ: "string", Str: "FULLRESYNC 8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb 0"}
}

func typ(args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.Value{Typ: "error", Str: "ERR wrong number of arguments for 'type' command"}
	}

	key := args[0].Bulk
	SETsMu.RLock()
	_, isString := SETs[key]
	SETsMu.RUnlock()

	STREAMsMu.RLock()
	_, isStream := STREAMs[key]
	STREAMsMu.RUnlock()

	if isString {
		return resp.Value{Typ: "string", Str: "string"}
	} else if isStream {
		return resp.Value{Typ: "string", Str: "stream"}
	}

	return resp.Value{Typ: "string", Str: "none"}
}

type Stream struct {
	id   string
	KVPs map[string]string
}

var STREAMs = map[string]Stream{}
var STREAMsMu = sync.RWMutex{}
var lastStreamID string = "0-0"

func xadd(args []resp.Value) resp.Value {
	if len(args) < 4 || len(args)%2 != 0 {
		return resp.Value{Typ: "error", Str: "ERR wrong number of arguments for 'xadd' command"}
	}

	newStreamID := tryGenarateStreamId(args[1].Bulk)
	
	if isLessThanOrEqual(newStreamID, lastStreamID) {
		if isLessThanOrEqual(newStreamID, "0-0") {
			return resp.Value{Typ: "error", Str: "ERR The ID specified in XADD must be greater than 0-0"}
		} else {
			return resp.Value{Typ: "error", Str: "ERR The ID specified in XADD is equal or smaller than the target stream top item"}
		}
	}

	streamKey := args[0].Bulk
	stream := Stream{
		id:   newStreamID,
		KVPs: make(map[string]string),
	}

	for i := 2; i < len(args); i += 2 {
		stream.KVPs[args[i].Bulk] = args[i+1].Bulk
	}

	STREAMsMu.Lock()
	STREAMs[streamKey] = stream
	STREAMsMu.Unlock()

	lastStreamID = stream.id
	return resp.Value{Typ: "bulk", Bulk: stream.id}
}

func tryGenarateStreamId(input string) string {
	inputSplit := strings.Split(input, "-")
	if len(inputSplit) == 2 && inputSplit[1] != "*" {
		return input
	}

	if inputSplit[0] == "*" {
		inputSplit[0] = strconv.FormatInt(time.Now().UnixMilli(), 10)
	}

	STREAMsMu.RLock()
	streams := STREAMs
	STREAMsMu.RUnlock()

	var sequence int64 = -1
	for _, stream := range streams {
		streamIDSplit := strings.Split(stream.id, "-")
		seq, err := strconv.ParseInt(streamIDSplit[1], 10, 64)

		if err != nil {
			return input
		}

		if streamIDSplit[0] == inputSplit[0] && seq > sequence {
			sequence = seq
		}
	}

	if inputSplit[0] == "0" && sequence == -1 {
		sequence = 0
	}

	return inputSplit[0] + "-" + strconv.FormatInt(sequence + 1, 10)
}

func isLessThanOrEqual(id1, id2 string) bool {
	ID1Split := strings.Split(id1, "-")
	ID2Split := strings.Split(id2, "-")
	
	if len(ID1Split) != 2 || len(ID2Split) != 2 {
		// Invalid Redis ID format
		return false
	}
	
	timestamp1, err1 := strconv.ParseInt(ID1Split[0], 10, 64)
	timestamp2, err2 := strconv.ParseInt(ID2Split[0], 10, 64)
	sequence1, err3 := strconv.ParseInt(ID1Split[1], 10, 64)
	sequence2, err4 := strconv.ParseInt(ID2Split[1], 10, 64)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return false
	}

	
	return timestamp1 < timestamp2 || (timestamp1 == timestamp2 && sequence1 <= sequence2)
}

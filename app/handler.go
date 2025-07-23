package main

import (
	"fmt"
	"slices"
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
	"WAIT":     wait,
	"TYPE":     typ,
	"XADD":     xadd,
	"XRANGE":   xrange,
	"XREAD":    xread,
	"INCR":     incr,
	"RPUSH":    rpush,
	"LRANGE":   lrange,
	"LPUSH":    lpush,
	"LLEN":     llen,
	"LPOP":     lpop,
	"BLPOP":    blpop,
}

var WriteCommands []string = []string{"SET", "XADD", "INCR", "RPUSH", "LPUSH", "LPOP", "BLPOP"}

func ping(args []resp.Value) resp.Value {
	return resp.Value{Typ: resp.STRING_TYPE, Str: "PONG"}
}

func echo(args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.Value{Typ: resp.ERROR_TYPE, Str: "ERR wrong number of args for 'echo' command"}
	}

	return resp.Value{Typ: resp.STRING_TYPE, Str: args[0].Bulk}
}

var SETs = map[string]string{}
var SETsMu = sync.RWMutex{}

func set(args []resp.Value) resp.Value {
	if len(args) != 2 && len(args) != 4 {
		return resp.Value{Typ: resp.ERROR_TYPE, Str: "ERR wrong number of args for 'set' command"}
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
			return resp.Value{Typ: resp.ERROR_TYPE, Str: "ERR syntax error"}
		}

		i64, err := strconv.ParseInt(args[3].Bulk, 10, 64)
		if err != nil {
			return resp.Value{Typ: resp.ERROR_TYPE, Str: "value is not an integer or out of range"}
		}

		go func() {
			time.Sleep(unit * time.Duration(i64))
			unset(key)
		}()
	}

	return resp.Value{Typ: resp.STRING_TYPE, Str: "OK"}
}

func unset(key string) {
	SETsMu.Lock()
	delete(SETs, key)
	SETsMu.Unlock()
}

func get(args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.Value{Typ: resp.ERROR_TYPE, Str: "ERR wrong number of args for 'get' command"}
	}

	key := args[0].Bulk
	SETsMu.RLock()
	val, ok := SETs[key]
	SETsMu.RUnlock()

	if !ok {
		return resp.Value{Typ: resp.NULL_TYPE}
	}
	return resp.Value{Typ: resp.BULK_TYPE, Bulk: val}
}

func config(args []resp.Value) resp.Value {
	if len(args) != 2 {
		return resp.Value{Typ: resp.ERROR_TYPE, Str: "ERR wrong number of arguments for 'config' command"}
	}

	switch args[0].Bulk {
	case "GET":
		return configGet(args[1:])
	default:
		return resp.Value{Typ: resp.ERROR_TYPE, Str: fmt.Sprintf("ERR unknown subcommand '%v'", args[0].Bulk)}
	}
}

func configGet(args []resp.Value) resp.Value {
	key := args[0].Bulk
	value, ok := server.configs[key]

	ret := resp.Value{Typ: resp.ARRAY_TYPE}
	if ok {
		ret.Array = append(ret.Array, resp.Value{Typ: resp.BULK_TYPE, Bulk: key}, resp.Value{Typ: resp.BULK_TYPE, Bulk: value})
	}

	return ret
}

func keys(args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.Value{Typ: resp.ERROR_TYPE, Str: "ERR wrong number of arguments for 'keys' command"}
	}

	ret := resp.Value{Typ: resp.ARRAY_TYPE}
	SETsMu.RLock()
	for key := range SETs {
		ret.Array = append(ret.Array, resp.Value{Typ: resp.BULK_TYPE, Bulk: key})
	}
	SETsMu.RUnlock()

	return ret
}

func info(args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.Value{Typ: resp.ERROR_TYPE, Str: "ERR wrong number of arguments for 'info' command"}
	}

	switch strings.ToUpper(args[0].Bulk) {
	case "REPLICATION":
		return infoReplication()
	default:
		return resp.Value{Typ: resp.BULK_TYPE}
	}
}

func infoReplication() resp.Value {
	ret := resp.Value{Typ: resp.BULK_TYPE}
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
		return resp.Value{Typ: resp.STRING_TYPE, Str: "OK"}
	}
}

func replconfgetack(args []resp.Value) resp.Value {
	if args[0].Bulk != "*" {
		return resp.Value{Typ: resp.ERROR_TYPE, Str: "Invalid GETACK parameter"}
	}

	offset := server.offset
	for _, slave := range server.slaves {
		offset += slave.offset
	}

	ret := resp.Value{Typ: resp.ARRAY_TYPE}
	ret.Array = append(ret.Array, resp.Value{Typ: resp.BULK_TYPE, Bulk: "REPLCONF"}, resp.Value{Typ: resp.BULK_TYPE, Bulk: "ACK"}, resp.Value{Typ: resp.BULK_TYPE, Bulk: strconv.Itoa(offset)})
	return ret
}

func psync(args []resp.Value) resp.Value {
	return resp.Value{Typ: resp.STRING_TYPE, Str: "FULLRESYNC 8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb 0"}
}

func wait(args []resp.Value) resp.Value {
	return resp.Value{Typ: resp.INTEGER_TYPE, Int: len(server.slaves)}
}

func typ(args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.Value{Typ: resp.ERROR_TYPE, Str: "ERR wrong number of arguments for 'type' command"}
	}

	key := args[0].Bulk
	SETsMu.RLock()
	_, isString := SETs[key]
	SETsMu.RUnlock()

	streams.mu.RLock()
	_, isStream := streams.entries[key]
	streams.mu.RUnlock()

	if isString {
		return resp.Value{Typ: resp.STRING_TYPE, Str: "string"}
	} else if isStream {
		return resp.Value{Typ: resp.STRING_TYPE, Str: "stream"}
	}

	return resp.Value{Typ: resp.STRING_TYPE, Str: "none"}
}

type Streams struct {
	entries     map[string]Stream
	mu          sync.RWMutex
	change      chan struct{}
	trackChange bool
}

type Stream struct {
	last    string
	entries []StreamEntry
}

type StreamEntry struct {
	id   string
	KVPs map[string]string
}

var streams Streams = Streams{
	entries:     make(map[string]Stream),
	mu:          sync.RWMutex{},
	change:      make(chan struct{}),
	trackChange: false,
}

func xadd(args []resp.Value) resp.Value {
	if len(args) < 4 || len(args)%2 != 0 {
		return resp.Value{Typ: resp.ERROR_TYPE, Str: "ERR wrong number of arguments for 'xadd' command"}
	}

	streamKey := args[0].Bulk
	streams.mu.RLock()
	stream, ok := streams.entries[streamKey]
	streams.mu.RUnlock()

	if !ok {
		stream = Stream{
			last:    "0-0",
			entries: make([]StreamEntry, 0),
		}
	}

	newStreamEntryID := tryGenarateStreamEntryId(args[1].Bulk, stream)

	if isLessThanOrEqual(newStreamEntryID, stream.last) {
		if isLessThanOrEqual(newStreamEntryID, "0-0") {
			return resp.Value{Typ: resp.ERROR_TYPE, Str: "ERR The ID specified in XADD must be greater than 0-0"}
		} else {
			return resp.Value{Typ: resp.ERROR_TYPE, Str: "ERR The ID specified in XADD is equal or smaller than the target stream top item"}
		}
	}

	entry := StreamEntry{
		id:   newStreamEntryID,
		KVPs: make(map[string]string),
	}

	for i := 2; i < len(args); i += 2 {
		entry.KVPs[args[i].Bulk] = args[i+1].Bulk
	}

	stream.entries = append(stream.entries, entry)
	stream.last = newStreamEntryID

	streams.mu.Lock()
	streams.entries[streamKey] = stream
	streams.mu.Unlock()

	if streams.trackChange {
		streams.change <- struct{}{}
	}

	return resp.Value{Typ: resp.BULK_TYPE, Bulk: entry.id}
}

func tryGenarateStreamEntryId(input string, stream Stream) string {
	inputSplit := strings.Split(input, "-")
	if len(inputSplit) == 2 && inputSplit[1] != "*" {
		return input
	}

	if inputSplit[0] == "*" {
		inputSplit[0] = strconv.FormatInt(time.Now().UnixMilli(), 10)
	}

	var sequence int64 = 0
	streamIDSplit := strings.Split(stream.last, "-")

	lastSeq, err := strconv.ParseInt(streamIDSplit[1], 10, 64)
	if err != nil {
		return input
	}

	if streamIDSplit[0] == inputSplit[0] {
		sequence = lastSeq + 1
	}

	return inputSplit[0] + "-" + strconv.FormatInt(sequence, 10)
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

func xrange(args []resp.Value) resp.Value {
	if len(args) != 3 {
		return resp.Value{Typ: resp.ERROR_TYPE, Str: "ERR wrong number of arguments for 'xrange' command"}
	}

	ret := resp.Value{Typ: resp.ARRAY_TYPE}
	streamKey := args[0].Bulk
	streams.mu.RLock()
	stream, ok := streams.entries[streamKey]
	streams.mu.RUnlock()

	if !ok {
		return resp.Value{Typ: resp.NULL_TYPE}
	}

	startVal, startSeq := xrangeFormatArgs(args[1].Bulk, stream)
	endVal, endSeq := xrangeFormatArgs(args[2].Bulk, stream)

	for _, entry := range stream.entries {
		entryId := strings.Split(entry.id, "-")
		seq, _ := strconv.ParseInt(entryId[1], 10, 64)

		if (entryId[0] > startVal || (entryId[0] == startVal && seq >= startSeq)) && (entryId[0] < endVal || (entryId[0] == endVal && seq <= endSeq)) {
			valsArr := resp.Value{Typ: resp.ARRAY_TYPE}
			for key, value := range entry.KVPs {
				valsArr.Array = append(valsArr.Array, resp.Value{Typ: resp.BULK_TYPE, Bulk: key}, resp.Value{Typ: resp.BULK_TYPE, Bulk: value})
			}

			respEntry := resp.Value{Typ: resp.ARRAY_TYPE}
			respEntry.Array = append(respEntry.Array, resp.Value{Typ: resp.BULK_TYPE, Bulk: entry.id}, valsArr)
			ret.Array = append(ret.Array, respEntry)
		}
	}

	if len(ret.Array) == 0 {
		return resp.Value{Typ: resp.NULL_TYPE}
	}

	return ret
}

func xrangeFormatArgs(arg string, stream Stream) (string, int64) {
	splitArg := strings.Split(arg, "-")
	if splitArg[0] == "-" {
		return "0", 0
	}

	if splitArg[0] == "+" {
		splitArg = strings.Split(stream.last, "-")
	}

	val := splitArg[0]
	var seq int64
	if len(splitArg) == 2 {
		seq, _ = strconv.ParseInt(splitArg[1], 10, 64)
	}

	return val, seq
}

func xread(args []resp.Value) resp.Value {
	searchData := args[1:]
	var blockDuration int = -1
	if args[0].Bulk == "block" {
		blockDuration, _ = strconv.Atoi(args[1].Bulk)
		searchData = args[3:]
	}

	median := len(searchData) / 2
	streamKeys := make([]string, 0)

	for i := 0; i+median < len(searchData); i++ {
		streamKeys = append(streamKeys, searchData[i].Bulk)
		if searchData[i+median].Bulk == "$" {
			streams.mu.RLock()
			stream, ok := streams.entries[searchData[i].Bulk]
			streams.mu.RUnlock()
			if ok {
				searchData[median+i].Bulk = stream.last
			} else {
				searchData[median+i].Bulk = "0-0"
			}
		}
	}

	if blockDuration == 0 {
		streams.trackChange = true
		<-streams.change
		streams.trackChange = false
	}

	time.Sleep(time.Duration(blockDuration) * time.Millisecond)

	ret := resp.Value{Typ: resp.ARRAY_TYPE}
	if len(streamKeys) == 0 {
		return resp.Value{Typ: resp.NULL_TYPE}
	}

	foundEntry := false
	for i, streamKey := range streamKeys {
		if median+i >= len(searchData) {
			return resp.Value{Typ: resp.ERROR_TYPE, Str: "ERR Unbalanced 'xread' list of streams: for each stream key an ID or '$' must be specified."}
		}

		streams.mu.RLock()
		stream, ok := streams.entries[streamKey]
		streams.mu.RUnlock()
		if !ok {
			continue
		}

		startVal, startSeq := xrangeFormatArgs(searchData[median+i].Bulk, stream)
		respStream := resp.Value{Typ: resp.ARRAY_TYPE}
		respStream.Array = append(respStream.Array, resp.Value{Typ: resp.BULK_TYPE, Bulk: streamKey})

		for _, entry := range stream.entries {
			entryId := strings.Split(entry.id, "-")
			seq, _ := strconv.ParseInt(entryId[1], 10, 64)

			if entryId[0] > startVal || (entryId[0] == startVal && seq > startSeq) {
				respEntries := resp.Value{Typ: resp.ARRAY_TYPE}
				respEntry := resp.Value{Typ: resp.ARRAY_TYPE}
				respEntry.Array = append(respEntry.Array, resp.Value{Typ: resp.BULK_TYPE, Bulk: entry.id})

				respKVPs := resp.Value{Typ: resp.ARRAY_TYPE}
				for key, value := range entry.KVPs {
					respKVPs.Array = append(respKVPs.Array, resp.Value{Typ: resp.BULK_TYPE, Bulk: key}, resp.Value{Typ: resp.BULK_TYPE, Bulk: value})
				}

				respEntry.Array = append(respEntry.Array, respKVPs)
				respEntries.Array = append(respEntries.Array, respEntry)
				respStream.Array = append(respStream.Array, respEntries)
				foundEntry = true
			}
		}

		ret.Array = append(ret.Array, respStream)
	}

	if !foundEntry {
		return resp.Value{Typ: resp.NULL_TYPE}
	}

	return ret
}

func incr(args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.Value{Typ: resp.ERROR_TYPE, Str: "ERR wrong number of arguments for 'incr' command"}
	}

	key := args[0].Bulk
	SETsMu.RLock()
	val, ok := SETs[key]
	SETsMu.RUnlock()

	if !ok {
		val = "0"
	}

	i, err := strconv.Atoi(val)
	if err != nil {
		return resp.Value{Typ: resp.ERROR_TYPE, Str: "ERR value is not an integer or out of range"}
	}

	SETsMu.Lock()
	SETs[key] = strconv.Itoa(i + 1)
	SETsMu.Unlock()

	return resp.Value{Typ: resp.INTEGER_TYPE, Int: i + 1}
}

func multi(queue *Queue) resp.Value {
	queue.active = true
	return resp.Value{Typ: resp.STRING_TYPE, Str: "OK"}
}

func exec(queue *Queue) resp.Value {
	if !queue.active {
		return resp.Value{Typ: resp.ERROR_TYPE, Str: "ERR EXEC without MULTI"}
	}

	ret := resp.Value{Typ: resp.ARRAY_TYPE}
	for _, item := range queue.items {
		command := strings.ToUpper(item.Array[0].Bulk)
		handler := Handlers[command]
		ret.Array = append(ret.Array, ExecuteCommand(handler, item.Array[1:]))
	}

	queue.active = false
	return ret
}

func discard(queue *Queue) resp.Value {
	if !queue.active {
		return resp.Value{Typ: resp.ERROR_TYPE, Str: "ERR DISCARD without MULTI"}
	}

	queue.active = false
	queue.items = nil
	return resp.Value{Typ: resp.STRING_TYPE, Str: "OK"}
}

type List struct {
	items []resp.Value
}

type Lists struct {
	lists map[string]List
	mu    sync.Mutex
	cond  *sync.Cond
}

var lists *Lists = newLists()

func newLists() *Lists {
	lists := &Lists{
		lists: make(map[string]List),
	}
	lists.cond = sync.NewCond(&lists.mu)
	return lists
}

func rpush(args []resp.Value) resp.Value {
	if len(args) < 2 {
		return resp.Value{Typ: resp.ERROR_TYPE, Str: "ERR wrong number of arguments for 'rpush' command"}
	}

	key := args[0].Bulk

	lists.mu.Lock()
	defer lists.mu.Unlock()

	list, ok := lists.lists[key]
	if !ok {
		list = List{}
	}

	list.items = append(list.items, args[1:]...)
	lists.lists[key] = list
	lists.cond.Signal()

	return resp.Value{Typ: resp.INTEGER_TYPE, Int: len(list.items)}
}

func lrange(args []resp.Value) resp.Value {
	if len(args) != 3 {
		return resp.Value{Typ: resp.ERROR_TYPE, Str: "ERR wrong number of arguments for 'lrange' command"}
	}

	key := args[0].Bulk
	lists.mu.Lock()
	defer lists.mu.Unlock()

	list, ok := lists.lists[key]
	if !ok {
		return resp.Value{Typ: resp.ARRAY_TYPE, Array: []resp.Value{}}
	}

	start, err := strconv.Atoi(args[1].Bulk)
	if err != nil {
		return resp.Value{Typ: resp.ERROR_TYPE, Str: err.Error()}
	}

	if start < 0 {
		start = max(len(list.items)+start, 0)
	}

	end, err := strconv.Atoi(args[2].Bulk)
	if err != nil {
		return resp.Value{Typ: resp.ERROR_TYPE, Str: err.Error()}
	}

	if end < 0 {
		end = max(len(list.items)+end, 0)
	}

	if end >= len(list.items) {
		end = len(list.items) - 1
	}

	if start >= len(list.items) || start > end {
		return resp.Value{Typ: resp.ARRAY_TYPE, Array: []resp.Value{}}
	}

	return resp.Value{Typ: resp.ARRAY_TYPE, Array: list.items[start : end+1]}
}

func lpush(args []resp.Value) resp.Value {
	if len(args) < 2 {
		return resp.Value{Typ: resp.ERROR_TYPE, Str: "ERR wrong number of arguments for 'lpush' command"}
	}
	key := args[0].Bulk

	lists.mu.Lock()
	defer lists.mu.Unlock()

	list, ok := lists.lists[key]
	if !ok {
		list = List{}
	}

	slices.Reverse(args[1:])
	list.items = append(args[1:], list.items...)
	lists.lists[key] = list
	lists.cond.Broadcast()

	return resp.Value{Typ: resp.INTEGER_TYPE, Int: len(list.items)}
}

func llen(args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.Value{Typ: resp.ERROR_TYPE, Str: "ERR wrong number of arguments for 'llen' command"}
	}

	key := args[0].Bulk
	lists.mu.Lock()
	defer lists.mu.Unlock()

	list := lists.lists[key]
	return resp.Value{Typ: resp.INTEGER_TYPE, Int: len(list.items)}
}

func lpop(args []resp.Value) resp.Value {
	if len(args) > 2 || len(args) < 1 {
		return resp.Value{Typ: resp.ERROR_TYPE, Str: "ERR wrong number of arguments for 'lpop' command"}
	}

	var (
		n   int = 1
		err error
	)

	if len(args) > 1 {
		fmt.Println(len(args))
		if n, err = strconv.Atoi(args[1].Bulk); err != nil {
			return resp.Value{Typ: resp.ERROR_TYPE, Str: err.Error()}
		}
	}

	if n < 0 {
		return resp.Value{Typ: resp.ERROR_TYPE, Str: "ERR value is out of range, must be positive"}
	}

	key := args[0].Bulk
	lists.mu.Lock()
	defer lists.mu.Unlock()

	list, ok := lists.lists[key]
	if !ok || len(list.items) == 0 {
		return resp.Value{Typ: resp.NULL_TYPE}
	}

	if n >= len(list.items) {
		n = len(list.items) - 1
	}

	items := list.items[:n]
	list.items = list.items[n:]
	lists.lists[key] = list

	if n > 1 {
		return resp.Value{Typ: resp.ARRAY_TYPE, Array: items}
	}

	return items[0]
}

func blpop(args []resp.Value) resp.Value {
	if len(args) != 2 {
		return resp.Value{Typ: resp.ERROR_TYPE, Str: "ERR wrong number of arguments for 'blpop' command"}
	}

	key := args[0].Bulk
	timeout, err := strconv.ParseFloat(args[1].Bulk, 32)
	if err != nil {
		return resp.Value{Typ: resp.ERROR_TYPE, Str: "ERR timeout is not a float or out of range"}
	}

	lists.mu.Lock()
	defer lists.mu.Unlock()

	var (
		list   List
		exists bool
	)

	for {
		list, exists = lists.lists[key]
		if exists && len(list.items) > 0 {
			break
		}

		if timeout == 0 {
			lists.cond.Wait()
		} else {
			timeoutCh := time.After(time.Duration(timeout * float64(time.Second)))
			waitCh := make(chan struct{})
			go func() {
				lists.cond.Wait()
				waitCh <- struct{}{}
			}()

			select {
			case <-waitCh:
			case <-timeoutCh:
				lists.cond.Signal()
				time.Sleep(100 * time.Millisecond) // Wait for the unlock
				return resp.Value{Typ: resp.NULL_TYPE}
			}
		}
	}

	item := list.items[0]
	list.items = list.items[1:]
	lists.lists[key] = list

	return resp.Value{Typ: resp.ARRAY_TYPE, Array: []resp.Value{args[0], item}}
}

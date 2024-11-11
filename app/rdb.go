package main

import (
	"encoding/binary"
	"os"
	"strconv"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

const (
	opCodeAUX          byte = 250
	opCodeRESIZEDB     byte = 251
	opCodeEXPIRETIMEMS byte = 252
	opCodeEXPIRETIME   byte = 253
	opCodeSELECTDB     byte = 254
	opCodeEOF          byte = 255
)

func sliceIndex(data []byte, sep byte) int {
	for i, b := range data {
		if b == sep {
			return i
		}
	}

	return -1
}

func parseTable(bytes []byte) []byte {
	start := sliceIndex(bytes, opCodeRESIZEDB)
	end := sliceIndex(bytes, opCodeEOF)
	return bytes[start+1 : end]
}

func readFile(path string) error {
	c, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if len(c) == 0 {
		return nil
	}

	type Expiry struct {
		Option byte
		UntilExpiry string
	}

	content := parseTable(c)
	var j byte = 2
	for i := byte(0); i < content[0]; i++ {
		expireOption := content[j]
		expiry := Expiry{Option: expireOption}

		if expireOption == opCodeEXPIRETIMEMS {
			i64 := binary.LittleEndian.Uint64(content[j+1 : j+9])
			unixTimeUTC := time.Unix(0, int64(i64) * int64(time.Millisecond)).UTC()
			j += 9
			expiry.UntilExpiry = strconv.Itoa(int(time.Until(unixTimeUTC).Seconds()))
		}

		if expireOption == opCodeEXPIRETIME {
			i32 := binary.LittleEndian.Uint32(content[j+1 : j+5])
			unixTimeUTC := time.Unix(int64(i32), 0)
			j += 5
			expiry.UntilExpiry = strconv.Itoa(int(time.Until(unixTimeUTC).Seconds()))
		}

		// Skip Value Type Byte
		j++

		kLen := content[j]
		key := content[j+1 : j+1+kLen]
		j += kLen + 1

		vLen := content[j]
		value := content[j+1 : j+1+vLen]
		j += vLen + 1

		args := make([]resp.Value, 2)
		args[0] = resp.Value{Typ: "bulk", Bulk: string(key)}
		args[1] = resp.Value{Typ: "bulk", Bulk: string(value)}

		switch expiry.Option {
		case opCodeEXPIRETIMEMS:
			args = append(args, resp.Value{Typ: "bulk", Bulk: "PX"}, resp.Value{Typ: "bulk", Bulk: expiry.UntilExpiry})
		case opCodeEXPIRETIME:
			args = append(args, resp.Value{Typ: "bulk", Bulk: "EX"}, resp.Value{Typ: "bulk", Bulk: expiry.UntilExpiry})
		}

		set(args)
	}
	

	return nil
}

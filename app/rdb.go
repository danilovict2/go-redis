package main

import (
	"os"
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

func readFile(path string) (map[string]string, error) {
	c, err := os.ReadFile(path)
	if err != nil {
		return map[string]string{}, err
	}

	if len(c) == 0 {
		return map[string]string{}, nil
	}

	content := parseTable(c)
	len := content[0]
	ret := make(map[string]string)
	var j byte = 2
	for i := byte(0); i < len; i++ {
		// Skip Value Type Byte
		j++

		kLen := content[j]
		key := content[j+1 : j+1+kLen]
		j += kLen + 1

		vLen := content[j]
		value := content[j+1 : j+1+vLen]
		j += vLen + 1
		
		ret[string(key)] = string(value)
	}

	return ret, nil
}

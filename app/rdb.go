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
	return bytes[start + 1 : end]
}

func readFile(path string) (string, error) {
	c, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	if len(c) == 0 {
		return "", nil
	}

	key := parseTable(c)
	str := key[4 : 4 + key[3]]
	return string(str), nil
}
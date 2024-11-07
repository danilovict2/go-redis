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

func readFile(path string) (map[string]string, error) {
	c, err := os.ReadFile(path)
	if err != nil {
		return map[string]string{}, err
	}

	if len(c) == 0 {
		return map[string]string{}, nil
	}

	content := parseTable(c)
	key := content[4 : 4 + content[3]]
	value := content[5 + content[3] : 5 + content[3] + content[4 + content[3]]]
	
	ret := make(map[string]string)
	ret[string(key)] = string(value)
	
	return ret, nil
}
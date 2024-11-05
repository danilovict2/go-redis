package main

import (
	"bufio"
	"io"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

func NewResp(rd io.Reader) *resp.Resp {
	return &resp.Resp{
		Reader: bufio.NewReader(rd),
	}
}

func NewWriter(w io.Writer) *resp.Writer {
	return &resp.Writer{
		Writer: w,
	}
}
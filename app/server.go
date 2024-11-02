package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	defer l.Close()

	for {
		conn, err := l.Accept()

		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go func() {
			defer conn.Close()
			for {
				resp := NewResp(conn)
				value, err := resp.Read()

				if errors.Is(err, io.EOF) {
					fmt.Println("Client closed the connections:", conn.RemoteAddr())
					break
				} else if err != nil {
					fmt.Println("Error while reading the message:", err)
					break
				}

				if value.typ != "array" {
					fmt.Println("Invalid request, expected array")
					break
				}
			
				if len(value.array) == 0 {
					fmt.Println("Invalid request, expected array length > 0")
					break
				}

				command := strings.ToUpper(value.array[0].bulk)
				handler, ok := Handlers[command]
				if !ok {
					fmt.Println("Invalid command: ", command)
					break
				}
				
				writer := NewWriter(conn)
				err = writer.Write(handler(value.array[1:]))
				if err != nil {
					fmt.Println("Error while writing the message:", err)
				}
			}
		}()
	}
}

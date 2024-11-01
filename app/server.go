package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
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
				buf := make([]byte, 128)
				_, err = conn.Read(buf)
				
				if errors.Is(err, io.EOF) {
					fmt.Println("Client closed the connections:", conn.RemoteAddr())
					break
				} else if err != nil {
					fmt.Println("Error while reading the message")
				}
	
				conn.Write([]byte("+PONG\r\n"))
			}
		}()
	}
}

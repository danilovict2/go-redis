package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	CONFIGsMu.RLock()
	port := CONFIGs["port"]
	CONFIGsMu.RUnlock()
	
	l, err := net.Listen("tcp", "0.0.0.0:" + port)
	if err != nil {
		fmt.Println("Failed to bind to port ", port)
		os.Exit(1)
	}

	defer l.Close()

	for {
		conn, err := l.Accept()

		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn)
	}
}

func init() {
	dir := flag.String("dir", "./", "the path to the directory where the RDB file is stored")
	dbfilename := flag.String("dbfilename", "dump.rdb", "the name of the RDB file")
	port := flag.String("port", "6379", "port number")
	replicaof := flag.String("replicaof", "", "start redis in replica mode")
	flag.Parse()

	CONFIGsMu.Lock()
	CONFIGs["dir"] = *dir;
	CONFIGs["dbfilename"] = *dbfilename
	CONFIGs["port"] = *port
	CONFIGs["replicaof"] = *replicaof
	CONFIGsMu.Unlock()

	path := *dir + "/" + *dbfilename
	err := readFile(path)
	if err != nil {
		fmt.Println("Error reading the RDB file:", err)
		return
	}
}

func handleConnection(conn net.Conn) {
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

		if value.Typ != "array" {
			fmt.Println("Invalid request, expected array")
			break
		}

		if len(value.Array) == 0 {
			fmt.Println("Invalid request, expected array length > 0")
			break
		}

		command := strings.ToUpper(value.Array[0].Bulk)
		handler, ok := Handlers[command]
		if !ok {
			fmt.Println("Invalid command: ", command)
			break
		}

		writer := NewWriter(conn)
		if err = writer.Write(handler(value.Array[1:])); err != nil {
			fmt.Println("Error while writing the message:", err)
		}
	}
}

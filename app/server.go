package main

import (
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
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
	}

	if *replicaof != "" {
		sendHandshake(*replicaof, *port)
	}
}

func sendHandshake(replicaof string, port string) {
	master := strings.Split(replicaof, " ")
	addres := fmt.Sprintf("%v:%v", master[0], master[1])
	conn, err := net.Dial("tcp", addres)
	if err != nil {
		fmt.Println("Couldn't connect to the master at ", addres)
		return
	}

	conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
	time.Sleep(1 * time.Second)
	conn.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$4\r\n" + port + "\r\n"))
	time.Sleep(1 * time.Second)
	conn.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n"))
	time.Sleep(1 * time.Second)
	conn.Write([]byte("*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n"))
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

		if command == "PSYNC" {
			emptyRDB, _ := hex.DecodeString("524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2")
			conn.Write(append([]byte(fmt.Sprintf("$%d\r\n", len(emptyRDB))), emptyRDB...))
		}
	}
}

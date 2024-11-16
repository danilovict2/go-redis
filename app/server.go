package main

import (
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

type ReplicaConfig struct {
	host string
	port string
}

type Server struct {
	configs     map[string]string
	replconf    ReplicaConfig
	slaves      []net.Conn
	listener    net.Listener
	broadcastch chan []byte
}

var server *Server

func main() {
	server = NewServer()
	initRDB(server.configs["dir"], server.configs["dbfilename"])
	server.Start()
}

func NewServer() *Server {
	dir := flag.String("dir", "./", "the path to the directory where the RDB file is stored")
	dbfilename := flag.String("dbfilename", "dump.rdb", "the name of the RDB file")
	port := flag.String("port", "6379", "port number")
	replicaof := flag.String("replicaof", "", "start redis in replica mode")
	flag.Parse()

	master := strings.Split(*replicaof, " ")
	replconf := ReplicaConfig{}
	if len(master) == 2 {
		replconf.host = master[0]
		replconf.port = master[1]
	}

	server := &Server{
		configs:     make(map[string]string),
		replconf:    replconf,
		broadcastch: make(chan []byte),
	}

	server.configs["port"] = *port
	server.configs["dir"] = *dir
	server.configs["dbfilename"] = *dbfilename

	return server
}

func initRDB(dir string, dbfilename string) {
	path := dir + "/" + dbfilename
	err := readFile(path)
	if err != nil {
		fmt.Println("Error reading the RDB file:", err)
	}
}

func (s *Server) Start() {
	l, err := net.Listen("tcp", "0.0.0.0:"+server.configs["port"])
	if err != nil {
		fmt.Println("Failed to bind to port ", server.configs["port"])
		os.Exit(1)
	}

	defer l.Close()
	s.listener = l

	if s.replconf.host != "" {
		go s.connectToMaster()
	} else {
		go s.propagateLoop()
	}

	s.Accept()
}

func (s *Server) Accept() {
	for {
		conn, err := s.listener.Accept()

		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go s.Handle(conn)
	}
}

func (s *Server) Handle(conn net.Conn) {
	defer conn.Close()
	resp := NewResp(conn)
	for {
		value, err := resp.Read()
		fmt.Println(value.Bulk)
		if errors.Is(err, io.EOF) {
			fmt.Println("Client closed the connections:", conn.RemoteAddr())
			break
		} else if err != nil {
			fmt.Println("Error while reading the message:", err)
			break
		}

		if value.Typ != "array" {
			fmt.Println("Invalid request, expected array")
			continue
		}

		if len(value.Array) < 1 {
			fmt.Println("Invalid request, expected array length > 0")
			continue
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
			continue
		}

		if command == "SET" || command == "REPLCONF" && len(s.slaves) > 0 {
			s.broadcastch <- value.Marshal()
		}

		if command == "PSYNC" {
			data, err := hex.DecodeString("524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2")
			if err != nil {
				fmt.Println("Error while decoding empty RDB file string")
				break
			}
			length := strconv.Itoa(len(data))
			_, err = conn.Write([]byte("$" + length + "\r\n" + string(data)))
			if err != nil {
				fmt.Println("Response with RDB file error")
				break
			}

			// Read the RDB to avoid errors
			res := make([]byte, 1024)
			_, _ = conn.Read(res)
			
			s.slaves = append(s.slaves, conn)
		}
	}
}

func (s *Server) connectToMaster() {
	res := make([]byte, 1024)
	conn, err := net.Dial("tcp", s.replconf.host+":"+s.replconf.port)
	if err != nil {
		fmt.Println("Couldn't connect to the master at ", s.replconf.host+":"+s.replconf.port)
		return
	}

	defer conn.Close()

	msg := resp.Value{Typ: "array", Array: []resp.Value{
		{
			Typ:  "bulk",
			Bulk: "ping",
		},
	}}

	_, err = conn.Write(msg.Marshal())
	if err != nil {
		fmt.Println("Error while writing handshake 1/3")
		os.Exit(1)
	}

	_, err = conn.Read(res)
	if err != nil {
		fmt.Println("Error while writing handshake 1/3")
		os.Exit(1)
	}

	msg = resp.Value{
		Typ: "array",
		Array: []resp.Value{
			{
				Typ:  "bulk",
				Bulk: "REPLCONF",
			},
			{
				Typ:  "bulk",
				Bulk: "listening-port",
			},
			{
				Typ:  "bulk",
				Bulk: s.configs["port"],
			},
		},
	}

	_, err = conn.Write(msg.Marshal())
	if err != nil {
		fmt.Println("Error while writing handshake 2a/3")
		os.Exit(1)
	}

	_, err = conn.Read(res)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Error while reading response 2a/3")
		os.Exit(1)
	}

	msg = resp.Value{
		Typ: "array",
		Array: []resp.Value{
			{
				Typ:  "bulk",
				Bulk: "REPLCONF",
			},
			{
				Typ:  "bulk",
				Bulk: "capa",
			},
			{
				Typ:  "bulk",
				Bulk: "psync2",
			},
		},
	}

	_, err = conn.Write(msg.Marshal())
	if err != nil {
		fmt.Println("Error while writing handshake 2b/3")
		os.Exit(1)
	}

	_, err = conn.Read(res)
	if err != nil {
		fmt.Println("Error while reading response 2b/3")
		os.Exit(1)
	}

	msg = resp.Value{
		Typ: "array",
		Array: []resp.Value{
			{
				Typ:  "bulk",
				Bulk: "PSYNC",
			},
			{
				Typ:  "bulk",
				Bulk: "?",
			},
			{
				Typ:  "bulk",
				Bulk: "-1",
			},
		},
	}

	_, err = conn.Write(msg.Marshal())
	if err != nil {
		fmt.Println("Error while writing handshake 3/3")
		os.Exit(1)
	}

	_, err = conn.Read(res)
	if err != nil {
		fmt.Println("Error while reading response 3/3")
		os.Exit(1)
	}
	s.Handle(conn)
}

func (s *Server) propagateLoop() {
	for {
		msg := <-s.broadcastch
		for _, server := range s.slaves {
			_, err := server.Write(msg)
			if err != nil {
				fmt.Println("Error broadcasting message to server:" + server.RemoteAddr().String())
			}
		}
	}
}

package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	listen, error := net.Listen("tcp", ":6379")
	if error != nil {
		log.Fatal("Cannot listen on :6379")
	}
	defer listen.Close()

	conn, error := listen.Accept()
	if error != nil {
		fmt.Println(error)
		os.Exit(1)
	}

	defer conn.Close()

	for {
		buf := make([]byte, 1024)
		conn.Read(buf)

		fmt.Println(string(buf))

		conn.Write([]byte("+OK\r\n"))
	}

}

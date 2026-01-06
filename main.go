package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	log.Print("reading config file")
	readConf("./redis.conf")
	listen, error := net.Listen("tcp", ":6379")
	if error != nil {
		log.Fatal("Cannot listen on :6379")
	}
	defer listen.Close()

	fmt.Println("Server is running on port 6379")

	conn, error := listen.Accept()
	if error != nil {
		fmt.Println(error)
		os.Exit(1)
	}

	fmt.Println("Connection Accepted")

	defer conn.Close()

	for {
		v := Value{typ: ARRAY}
		v.readArray(conn)
		handle(conn, &v)
	}
}

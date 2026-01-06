package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
)

type ValueType string

const (
	ARRAY  ValueType = "*"
	BULK   ValueType = "$"
	STRING ValueType = "+"
	ERROR  ValueType = "-"
	NULL   ValueType = ""
)

type Value struct {
	typ   ValueType
	bulk  string
	str   string
	array []Value
	err   ValueType
	null  ValueType
}

func (v *Value) readArray(reader io.Reader) {
	buf := make([]byte, 4)
	reader.Read(buf)
	arrLen, err := strconv.Atoi(string(buf[1]))
	if err != nil {
		fmt.Println(err)
		return
	}

	for range arrLen {
		bulk := v.readBulk(reader)
		v.array = append(v.array, bulk)
	}
}

func (v *Value) readBulk(reader io.Reader) Value {

	buf := make([]byte, 4)
	reader.Read(buf)

	n, err := strconv.Atoi(string(buf[1]))
	if err != nil {
		fmt.Println(err)
		return Value{}
	}

	bulkBuf := make([]byte, n+2)
	reader.Read(bulkBuf)

	bulk := string(bulkBuf[:n])
	return Value{typ: BULK, bulk: bulk}
}

func main() {
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

type Handler func(*Value) *Value

var Handlers = map[string]Handler{
	"COMMAND": command,
	"GET":     get,
	"SET":     set,
}

type Database struct {
	store map[string]string
	mu    sync.RWMutex
}

func NewDatabase() *Database {
	return &Database{
		store: map[string]string{},
		mu:    sync.RWMutex{},
	}
}

var DB = NewDatabase()

func handle(conn net.Conn, v *Value) {
	cmd := v.array[0].bulk
	handler, ok := Handlers[cmd]
	if !ok {
		fmt.Println("Invalid command ", cmd)
		return
	}

	reply := handler(v)
	w := NewWriter(conn)
	w.Write(reply)
}

func get(v *Value) *Value {
	args := v.array[1:]
	if len(args) != 1 {
		return &Value{typ: ERROR, err: "ERR invalid no of args for 'GET"}
	}
	name := args[0].bulk
	DB.mu.RLock()
	val, ok := DB.store[name]
	DB.mu.RUnlock()
	if !ok {
		return &Value{typ: NULL}
	}

	return &Value{typ: BULK, bulk: val}
}

func set(v *Value) *Value {
	args := v.array[1:]
	if len(args) != 2 {
		return &Value{typ: ERROR, err: "ERR invalid no of args for 'SET'"}
	}

	key := args[0].bulk
	val := args[1].bulk
	DB.mu.Lock()
	DB.store[key] = val
	DB.mu.Unlock()

	return &Value{typ: STRING, str: "OK"}
}

func command(v *Value) *Value {
	return &Value{typ: STRING, str: "OK"}
}

type Writer struct {
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: bufio.NewWriter(w)}
}

func (w *Writer) Write(v *Value) {
	var reply string
	switch v.typ {
	case STRING:
		reply = fmt.Sprintf("%s%s\r\n", v.typ, v.str)
	case BULK:
		reply = fmt.Sprintf("%s%d\r\n%s\r\n", v.typ, len(v.bulk), v.bulk)
	case ERROR:
		reply = fmt.Sprintf("%s%s\r\n", v.typ, v.err)
	case NULL:
		reply = "$-1\r\n"
	}

	w.writer.Write([]byte(reply))
	w.writer.(*bufio.Writer).Flush()
}

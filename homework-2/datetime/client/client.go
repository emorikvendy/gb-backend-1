package main

import (
	"io"
	"log"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	defer func(conn net.Conn) {
		err2 := conn.Close()
		if err2 != nil {
			log.Printf("Error while closing connection: %v", err2)
		}
	}(conn)
	buf := make([]byte, 256)

	for {
		_, err2 := conn.Read(buf)
		if err2 == io.EOF {
			break
		}
		if err2 != nil {
			log.Printf("Error while reading from connection: %v", err2)
			break
		}

		log.Printf("Read from server: %s", string(buf))
	}
}

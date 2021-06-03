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
		err := conn.Close()
		if err != nil {
			log.Printf("Error while closing connection: %v", err)
		}
	}(conn)
	buf := make([]byte, 256)

	for {
		_, err := conn.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error while reading from connection: %v", err)
			break
		}

		log.Printf("Read from server: %s", string(buf))
	}
}

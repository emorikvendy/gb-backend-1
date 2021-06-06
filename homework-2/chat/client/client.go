package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Printf("Error while closing connection: %v", err)
		}
	}()
	fmt.Println("Enter your name")
	var name string
	_, err = fmt.Scanln(&name)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(conn, name)
	go func() {
		_, err := io.Copy(os.Stdout, conn)
		if err != nil {
			log.Printf("Error while copying from connection: %v", err)
		}

	}()
	_, err = io.Copy(conn, os.Stdin)
	if err != nil {
		log.Printf("Error while copying to connection: %v", err)
	} // until you send ^Z
	fmt.Printf("%s: exit", conn.LocalAddr())
}

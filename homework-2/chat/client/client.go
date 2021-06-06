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
		err2 := conn.Close()
		if err2 != nil {
			log.Printf("Error while closing connection: %v", err2)
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
		_, err2 := io.Copy(os.Stdout, conn)
		if err2 != nil {
			log.Printf("Error while copying from connection: %v", err2)
		}
	}()
	_, err = io.Copy(conn, os.Stdin)
	if err != nil {
		log.Printf("Error while copying to connection: %v", err)
	} // until you send ^Z
	fmt.Printf("%s: exit", conn.LocalAddr())
}

package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	connections = NewConnSet()
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Server is ready to listen requests")
	cancel := make(chan struct{}, 1)
	wg := sync.WaitGroup{}
	defer close(cancel)
	wg.Add(1)
	go watchSignals(cancel, &wg)
	go closeListener(listener, cancel, &wg)
	go broadcaster(cancel)
	for {
		select {
		case <-cancel:
			cancel <- struct{}{}
			log.Print("loop finished")
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Error from connection: %v", err)
				continue
			}
			connections.Add(conn, struct{}{})
			wg.Add(1)
			go handleConn(conn, cancel, &wg)
		}
	}
}
func broadcaster(cancel chan struct{}) {
	for {
		select {
		case <-cancel:
			cancel <- struct{}{}
			log.Println("broadcaster stopped")
			return
		default:
			var msg string
			fmt.Scanln(&msg)
			connections.Range(func(conn net.Conn) {
				log.Printf("writing %s to %v", msg, conn)
				_, err := io.WriteString(conn, msg)
				if err != nil {
					log.Printf("Error while writing to connection: %v", err)
				}
			})
		}
	}
}

func closeListener(listener net.Listener, cancel chan struct{}, wg *sync.WaitGroup) {
	<-cancel
	log.Println("closeListener started")
	connections.Range(func(conn net.Conn) {
		delete(connections.m, conn)
		err := conn.Close()
		if err != nil {
			log.Printf("Error while closing connection: %v", err)
		}
	})
	cancel <- struct{}{}
	log.Print("All connections closed")
	wg.Wait()
	err := listener.Close()
	if err != nil {
		log.Printf("Error while closing listener: %v", err)
	}
}

func watchSignals(cancel chan struct{}, wg *sync.WaitGroup) {
	osSignalChan := make(chan os.Signal, 1)
	signal.Notify(osSignalChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-osSignalChan
	log.Printf("got signal %+v", sig)
	cancel <- struct{}{}
	wg.Done()
}

func handleConn(conn net.Conn, cancel chan struct{}, wg *sync.WaitGroup) {
	defer func() {
		connections.Delete(conn)
		err := conn.Close()
		if err != nil {
			log.Printf("Error while closing connection: %v", err)
		}
		wg.Done()
	}()
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-cancel:
			_, err := io.WriteString(conn, "Server is stopping")
			if err != nil {
				log.Printf("Error while writing to connection: %v", err)
			}
			cancel <- struct{}{}
			return
		case <-ticker.C:
			_, err := io.WriteString(conn, time.Now().Format("15:04:05"))
			if err != nil {
				return
			}
		}
	}
}

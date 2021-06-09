package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type client chan<- string

var (
	entering    = make(chan client)
	leaving     = make(chan client)
	messages    = make(chan string)
	connections = NewConnSet()
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	cancel := make(chan struct{}, 1)
	wg := &sync.WaitGroup{}
	defer close(cancel)
	wg.Add(1)
	go watchSignals(cancel, wg)
	go closeListener(listener, cancel, wg)

	go broadcaster()
	for {
		select {
		case <-cancel:
			cancel <- struct{}{}
			log.Print("loop finished")
			return
		default:
			conn, err := listener.Accept()
			connections.Add(conn, struct{}{})
			if err != nil {
				log.Print(err)
				continue
			}
			wg.Add(1)
			go handleConn(conn, cancel, wg)
			log.Printf("Connection added %v", conn)
		}
	}
}

func broadcaster() {
	clients := make(map[client]struct{})
	for {
		select {
		case msg := <-messages:
			for cli := range clients {
				cli <- msg
			}
			log.Printf("Message broadcasted: %s", msg)

		case cli := <-entering:
			clients[cli] = struct{}{}

		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}
func closeListener(listener net.Listener, cancel chan struct{}, wg *sync.WaitGroup) {
	<-cancel
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
	log.Print("watchSignals finished")
}

func handleConn(conn net.Conn, cancel chan struct{}, wg *sync.WaitGroup) {
	defer func() {
		log.Printf("handleConn defer for connection %v", conn)
		if connections.Has(conn) {
			connections.Delete(conn)
			err := conn.Close()
			if err != nil {
				log.Printf("Error while closing connection: %v", err)
			}
			log.Printf("Connection closed %v", conn)
		}
		log.Printf("handleConn stopped for connection %v", conn)
		wg.Done()
	}()
	ctx, cancelFunc := context.WithCancel(context.Background())
	ch := make(chan string)
	go clientWriter(ctx, conn, ch)

	//who := conn.RemoteAddr().String()
	//ch <- "You are " + who
	//messages <- who + " has arrived"
	//entering <- ch
	var who string

	input := bufio.NewScanner(conn)
	first := true
LOOP:
	for {
		select {
		case <-cancel:
			cancel <- struct{}{}
			break LOOP
		default:
			if first {
				ch <- "Enter your name"
			}
			if ok := input.Scan(); !ok {
				break LOOP
			}
			if first {
				who = input.Text()
				ch <- "You are " + who
				messages <- who + " has arrived"
				entering <- ch
				first = false
			} else {
				messages <- who + ": " + input.Text()
			}
		}
	}
	leaving <- ch
	messages <- who + " has left"
	cancelFunc()
}

func clientWriter(ctx context.Context, conn net.Conn, ch <-chan string) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("clientWriter stopped for connection %v", conn)
			return
		case msg := <-ch:
			fmt.Fprintln(conn, msg)
		}
	}
}

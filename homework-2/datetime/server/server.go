package main

import (
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
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
	go closeListener(listener, &wg)
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
			wg.Add(1)
			go handleConn(conn, cancel, &wg)
		}
	}
}

func closeListener(listener net.Listener, wg *sync.WaitGroup) {
	wg.Wait()
	err := listener.Close()
	if err != nil {
		log.Printf("Error while closing listener: %v", err)
	}
}

func watchSignals(cancel chan struct{}, wg *sync.WaitGroup) {
	osSignalChan := make(chan os.Signal)
	signal.Notify(osSignalChan, syscall.SIGINT, syscall.SIGTERM)
	<-osSignalChan
	cancel <- struct{}{}
	wg.Done()
}

func handleConn(c net.Conn, cancel chan struct{}, wg *sync.WaitGroup) {
	defer func() {
		err := c.Close()
		if err != nil {
			log.Printf("Error while closing connection: %v", err)
		}
		wg.Done()
	}()
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-cancel:
			io.WriteString(c, "Server is stopping")
			cancel <- struct{}{}
			return
		case <-ticker.C:
			_, err := io.WriteString(c, time.Now().Format("15:04:05"))
			if err != nil {
				return
			}
		}
	}
}

// Filename: main.go
package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

const (
	MaxMessageSize   = 1024
	InactivityPeriod = 30 * time.Second
)

func main() {
	port := flag.String("port", "4000", "Port number")
	flag.Parse()

	addr := ":" + *port
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server is listening on", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Connection error:", err)
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	ip := strings.Split(clientAddr, ":")[0]

	logFileName := ip + ".log"
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Could not open log file:", err)
		return
	}
	defer logFile.Close()

	logTime := time.Now().Format(time.RFC3339)
	fmt.Printf("[%s] %s connected\n", logTime, clientAddr)
	logFile.WriteString("[" + logTime + "] Connected\n")

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, MaxMessageSize), MaxMessageSize)

	timer := time.NewTimer(InactivityPeriod)
	defer timer.Stop()

	msgChan := make(chan string)

	go func() {
		for scanner.Scan() {
			msg := scanner.Text()
			msgChan <- msg
		}
		close(msgChan)
	}()

	for {
		select {
		case <-timer.C:
			conn.Write([]byte("Disconnected: Inactivity timeout\n"))
			return
		case msg, ok := <-msgChan:
			if !ok {
				fmt.Printf("[%s] %s disconnected\n", time.Now().Format(time.RFC3339), clientAddr)
				logFile.WriteString("[" + time.Now().Format(time.RFC3339) + "] Disconnected\n")
				return
			}
			timer.Reset(InactivityPeriod)

			msg = strings.TrimSpace(msg)
			if len(msg) > MaxMessageSize {
				msg = msg[:MaxMessageSize]
			}

			logFile.WriteString("[" + time.Now().Format(time.RFC3339) + "] " + msg + "\n")

			if msg == "" {
				conn.Write([]byte("Say something. . .\n"))
			} else if msg == "hello" {
				conn.Write([]byte("Hi there!\n"))
			} else if msg == "bye" {
				conn.Write([]byte("Goodbye!\n"))
				return
			} else if msg == "/quit" {
				conn.Write([]byte("Connection closed.\n"))
				return
			} else if msg == "/time" {
				conn.Write([]byte(time.Now().Format(time.RFC1123) + "\n"))
			} else if strings.HasPrefix(msg, "/echo ") {
				conn.Write([]byte(msg[6:] + "\n"))
			} else {
				conn.Write([]byte(msg + "\n"))
			}
		}
	}
}

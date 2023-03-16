// file main.go
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

var rooms = make(map[string]*room)

func main() {
	var port int
	flag.IntVar(&port, "port", 0, "port number for the server")
	flag.Parse()

	//case of no flag
	if port == 0 {
		fmt.Println("No 'port' argument passed, attempting to use network.config file...")
		file, err := os.Open("network.config")
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		if scanner.Scan() {
			port, err = strconv.Atoi(strings.TrimSpace(scanner.Text()))
			if err != nil {
				log.Fatal("Invalid port number in network.config")
			}
		}
	}

	// Check if port is available & Start the server
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Fatal("Port is not available: ", err)
	}
	defer ln.Close()

	fmt.Println("Server is running at port: ", port)

	//lobby room
	lobby := &room{
		name:    "lobby",
		clients: make(map[*client]struct{}),
	}
	rooms["lobby"] = lobby

	//client comms
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		conn.Write([]byte("/roulette-server\n"))
		go func(conn net.Conn) {
			c := &client{
				conn:     conn,
				command:  make(chan string),
				badFaith: 0,
			}
			c.read()
		}(conn)
	}
}

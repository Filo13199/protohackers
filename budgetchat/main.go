package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"slices"
	"strings"
	"sync"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var clientCount int

type ChatClient struct {
	Name string
	Conn *net.TCPConn
	Id   primitive.ObjectID
}

var clients []ChatClient

func main() {
	// initialize tcp server
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{Port: 17500})
	if err != nil {
		panic(err)
	}

	mu := sync.Mutex{}
	defer listener.Close()
	clients = []ChatClient{}
	clientNames := []string{}
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			panic(err)
		}

		fmt.Println("recieved connection !, prompting for name")
		clientCount++

		sentBytes, err := conn.Write([]byte("What should i call you??\n"))
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(sentBytes)
		reader := bufio.NewReader(conn)

		name, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			// Handle potential errors like EOF (client disconnected) or read timeouts
			log.Printf("Error reading from connection: %v", err)
			return
		}

		mu.Lock()

		client := ChatClient{
			Name: string(name),
			Conn: conn,
			Id:   primitive.NewObjectID(),
		}
		clients = append(clients, client)

		for i := range clients {
			if clients[i].Id != client.Id {
				_, err = clients[i].Conn.Write([]byte(fmt.Sprintf("* %s has entered the room\n", name)))
				if err != nil {
					log.Fatal(err)
				}
			}
		}

		_, err = conn.Write([]byte(fmt.Sprintf("* the room contains: %s\n", strings.Join(clientNames, ", "))))
		if err != nil {
			log.Fatal(err)
		}

		clientNames = append(clientNames, string(name))
		mu.Unlock()
		go chat(conn, client, &mu)
	}
}

func chat(conn *net.TCPConn, client ChatClient, mu *sync.Mutex) {
	defer func() {
		for i := range clients {
			if clients[i].Id != client.Id {
				_, err := clients[i].Conn.Write([]byte(fmt.Sprintf("* %s has left the room\n", client.Name)))
				if err != nil {
					log.Fatal(err)
				}
			}
		}
		mu.Lock()
		clients = slices.DeleteFunc(clients, func(c ChatClient) bool {
			return c.Id == client.Id
		})
		mu.Unlock()
		conn.Close()
	}()

	reader := bufio.NewReader(conn)
	// handle connection

	for {
		msg, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			log.Fatal(err)
		}
		mu.Lock()
		for i := range clients {
			if clients[i].Id != client.Id {
				_, err = clients[i].Conn.Write([]byte(fmt.Sprintf("[%s] %s\n", client.Name, string(msg))))
				if err != nil {
					log.Fatal(err)
				}
			}
		}
		mu.Unlock()
	}
}

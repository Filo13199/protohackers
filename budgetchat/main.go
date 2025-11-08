package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"regexp"
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
	rxgx := regexp.MustCompile("^[a-zA-Z0-9]*")
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

		match := rxgx.MatchString(name)
		name = strings.TrimRight(name, "\r\n")
		if !match || len(name) == 0 || slices.Contains(clientNames, name) {
			fmt.Printf("invalid name [%s]", name)
			_, err = conn.Write([]byte("invalid name !\n"))
			if err != nil {
				log.Fatal(err)
			}

			conn.Close()
			continue
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
				writeFunc(&clients[i], []byte(fmt.Sprintf("* %s has entered the room\n", name)))
			}
		}

		writeFunc(&client, []byte(fmt.Sprintf("* The room contains: %s \n", strings.Join(clientNames, ", "))))
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
				writeFunc(&clients[i], []byte(fmt.Sprintf("* %s has left the room\n", client.Name)))
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

		fmt.Println("message before trimming", msg)
		msg = strings.TrimRight(msg, "\r\n")
		fmt.Println("message after trimming", msg)
		content := "[" + client.Name + "] " + msg + "\n"

		fmt.Println("client isss ", client)
		mu.Lock()
		for i := range clients {
			if clients[i].Id != client.Id {
				writeFunc(&clients[i], []byte(content))
			}
		}
		mu.Unlock()
	}
}

func writeFunc(client *ChatClient, b []byte) {
	fmt.Printf("Sending message to %s, content= [%s], newlines count = %d", client.Name, string(b), bytes.Count(b, []byte{'\n'}))
	_, err := client.Conn.Write(b)
	if err != nil {
		log.Fatal(err)
	}
}

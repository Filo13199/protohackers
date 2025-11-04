package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"slices"
)

func main() {
	// initialize tcp server
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{Port: 17500})
	if err != nil {
		panic(err)
	}

	defer listener.Close()

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			panic(err)
		}
		fmt.Println("Accepted connection !")
		go meansToAnEnd(conn)
	}
}

type Tuple struct {
	Price     int32
	Timestamp int32
}

func meansToAnEnd(conn *net.TCPConn) {
	defer conn.Close()
	// handle connection
	data := []Tuple{}
	for {
		msg := make([]byte, 0, 9)
		_, err := conn.Read(msg)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("recieved message !, [%s], \n", string(msg))
		if len(msg) == 0 {
			break
		}

		op := string(msg[0])
		switch op {
		case "I":
			timestamp := binary.BigEndian.Uint32(msg[1:5])
			price := binary.BigEndian.Uint32(msg[5:])
			data = append(data, Tuple{
				Timestamp: int32(timestamp),
				Price:     int32(price),
			})
			slices.SortFunc(data, func(a, b Tuple) int {
				if a.Timestamp > b.Timestamp {
					return 1
				} else if b.Timestamp > a.Timestamp {
					return -1
				}
				return 0
			})
		case "Q":
			t1 := binary.BigEndian.Uint32(msg[1:5])
			t2 := binary.BigEndian.Uint32(msg[5:])
			i1 := int(math.Max(0, float64(slices.IndexFunc(data, func(t Tuple) bool {
				return t.Timestamp > int32(t1)
			}))))

			i2 := int(math.Max(float64(slices.IndexFunc(data, func(t Tuple) bool {
				return t.Timestamp > int32(t2)
			})), float64(len(data)-1)))

			sum := int32(0)
			count := 0
			for i := i1; i < i2; i++ {
				sum += data[i].Price
				count++
			}
		default:
			return
		}

	}
}

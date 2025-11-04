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
		msg := make([]byte, 9)
		n, err := conn.Read(msg)
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Println(n)
		fmt.Printf("recieved message !, [%s], \n", string(msg))
		if len(msg) == 0 {
			break
		}

		op := string(msg[0])
		operand1 := binary.BigEndian.Uint32(msg[1:5])
		operand2 := binary.BigEndian.Uint32(msg[5:])
		switch op {
		case "I":
			data = append(data, Tuple{
				Timestamp: int32(operand1),
				Price:     int32(operand2),
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
			i1 := int(math.Max(0, float64(slices.IndexFunc(data, func(t Tuple) bool {
				return t.Timestamp > int32(operand1)
			}))))

			i2 := int(math.Max(float64(slices.IndexFunc(data, func(t Tuple) bool {
				return t.Timestamp > int32(operand2)
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

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net"
	"slices"
)

var clientCount int

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
		clientCount++
		go meansToAnEnd(conn, clientCount)
	}
}

type Tuple struct {
	Price     int32
	Timestamp int32
}

func meansToAnEnd(conn *net.TCPConn, clientId int) {
	defer conn.Close()
	// handle connection
	data := []Tuple{}

	for {
		msg := make([]byte, 9)
		_, err := io.ReadFull(conn, msg)
		if err != nil {
			fmt.Println(err)
			break
		}
		if len(msg) == 0 {
			break
		}

		op := string(msg[0])
		operand1 := int32(binary.BigEndian.Uint32(msg[1:5]))
		operand2 := int32(binary.BigEndian.Uint32(msg[5:]))
		fmt.Printf("CLIENT #%d [%s	%d	%d]\n", clientId, op, operand1, operand2)
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

			fmt.Println(data)
		case "Q":
			i1 := int(math.Max(0, float64(slices.IndexFunc(data, func(t Tuple) bool {
				return t.Timestamp >= int32(operand1)
			}))))

			i2 := int(math.Max(float64(slices.IndexFunc(data, func(t Tuple) bool {
				return t.Timestamp > int32(operand2)
			})), float64(len(data)-1)))

			if i2 > 0 {
				i2--
			}

			sum := int32(0)
			count := int32(0)
			for i := i1; i <= i2; i++ {
				sum += data[i].Price
				count++
			}

			buf := new(bytes.Buffer)

			// Write the int32 to the buffer using BigEndian byte order
			// You can also use binary.LittleEndian or binary.NativeEndian
			avg := int32(0)
			if count > 0 {
				avg = sum / count
			}
			err := binary.Write(buf, binary.BigEndian, avg)
			if err != nil {
				fmt.Println("binary.Write failed:", err)
				return
			}
			conn.Write(buf.Bytes())
		default:
			return
		}

	}
}

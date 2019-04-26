package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"protocol"
	"time"
)

type dataPacket struct {
	Msgid   int
	UID     int
	Cusid   int
	Gameid  int
	Content string
}

type cusinfo struct {
	connection net.Conn
	count      int
	cusid      int
}

type userinfo struct {
	connection net.Conn
	cusConn    net.Conn
	cusid      int
	uid        int
}

var (
	cuspool = make(map[int]cusinfo)
	uidpool = make(map[int]userinfo)
)

func main() {
	service := ":7777"

	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		go handleConnection(conn, 5)
	}
}

func handleConnection(conn net.Conn, timeout int) {
	buffer := make([]byte, 2048)

	tmpBuffer := make([]byte, 0)
	readerChannel := make(chan []byte, 16)
	go reader(readerChannel, conn)
	for {
		n, err := conn.Read(buffer)

		if err != nil {
			logInfo(err)
			return
		}
		// Data := (buffer[:n])
		// messnager := make(chan byte)
		// //心跳计时
		// go heartBeating(conn, messnager, timeout)
		// //检测每次Client是否有数据传来
		// go gravelChannel(Data, messnager)
		tmpBuffer = protocol.Unpack(append(tmpBuffer, buffer[:n]...), readerChannel)
	}
}

func heartBeating(conn net.Conn, readerChannel chan byte, timeout int) {
	select {
	case fk := <-readerChannel:
		fmt.Println(fk)
		// fmt.Println(conn.RemoteAddr().String(), "receive data string:", string(fk))
		conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
		break
	case <-time.After(time.Second * 5):
		fmt.Println("It's really weird to get Nothing!!!")
		conn.Close()
	}
}

func gravelChannel(n []byte, mess chan byte) {
	for _, v := range n {
		mess <- v
	}
	close(mess)
}

func reader(readerChannel chan []byte, conn net.Conn) {
	for {
		select {
		case data := <-readerChannel:
			rountine(data, conn)
		}
	}
}

func rountine(b []byte, conn net.Conn) {
	var data dataPacket
	json.Unmarshal(b, &data)
	switch data.Msgid {
	case 1: //玩家连入，分配客服
		var cusConn net.Conn
		var cusid int
		if _, ok := uidpool[data.UID]; ok {
			cusid = uidpool[data.UID].cusid
			cusConn = uidpool[data.UID].cusConn
		} else {
			cusid = 0
			countTmp := 0
			for k, v := range cuspool {
				if countTmp == 0 {
					countTmp = v.count
					cusid = k
					cusConn = v.connection
				} else {
					if v.count < countTmp {
						countTmp = v.count
						cusid = k
						cusConn = v.connection
					}
				}
			}
		}

		if cusConn != nil {
			fmt.Println("User ", data.UID, " connect ok,the cus is", cusid)
			uidpool[data.UID] = userinfo{
				connection: conn,
				cusConn:    cusConn,
				cusid:      cusid,
				uid:        data.UID,
			}
			userinfo := dataPacket{
				Msgid:   1,
				UID:     data.UID,
				Cusid:   cusid,
				Gameid:  1,
				Content: "ff",
			}
			packetBytes, err := json.Marshal(userinfo)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			conn.Write(protocol.Packet(packetBytes))
			fmt.Println("Send cusinfo over")
		} else {
			logg("No Cusserver")
			return
		}
	case 2: //客服连入
		if _, ok := cuspool[data.Cusid]; ok {
			cuspool[data.Cusid].connection.Close()
		}
		cuspool[data.Cusid] = cusinfo{
			connection: conn,
			count:      0,
			cusid:      data.Cusid,
		}
		fmt.Println("Cusserver ", data.Cusid, " connect ok")
	case 3: //玩家发送消息
		fmt.Println("case 3")
	case 4: //客服发送消息
		fmt.Println("case 4")

	default:
		fmt.Println("default")
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func logInfo(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Log info: %s", err.Error())
	}
}

func logg(str interface{}) {
	fmt.Println(str)
}

package protocol

import (
	"bytes"
	"encoding/binary"
)

const (
	//TCPHeader 包头
	TCPHeader = "TCPHEADER"
	//TCPHeaderLen 包头长度
	TCPHeaderLen = 9
	//TCPDataLen 数据长度
	TCPDataLen = 4
)

//Packet 封包
func Packet(message []byte) []byte {
	return append(append([]byte(TCPHeader), IntToBytes(len(message))...), message...)
}

//Unpack 解包
func Unpack(buffer []byte, readerChannel chan []byte) []byte {
	length := len(buffer)

	var i int
	for i = 0; i < length; i = i + 1 {
		if length < i+TCPHeaderLen+TCPDataLen {
			break
		}
		if string(buffer[i:i+TCPHeaderLen]) == TCPHeader {
			messageLength := BytesToInt(buffer[i+TCPHeaderLen : i+TCPHeaderLen+TCPDataLen])
			if length < i+TCPHeaderLen+TCPDataLen+messageLength {
				break
			}
			data := buffer[i+TCPHeaderLen+TCPDataLen : i+TCPHeaderLen+TCPDataLen+messageLength]
			readerChannel <- data

			i += TCPHeaderLen + TCPDataLen + messageLength - 1
		}
	}

	if i == length {
		return make([]byte, 0)
	}
	return buffer[i:]
}

//IntToBytes 整形转换成字节
func IntToBytes(n int) []byte {
	x := int32(n)

	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

//BytesToInt 字节转换成整形
func BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)

	var x int32
	binary.Read(bytesBuffer, binary.BigEndian, &x)

	return int(x)
}

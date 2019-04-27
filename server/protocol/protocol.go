package protocol

import (
	"bytes"
	"encoding/binary"
)

const (
	ConstHeader  = "gitproxy"
	ConstPLength = 4 // 协议id
	ConstMLength = 4 // 消息长度

	Heartbeat = 11
	Subscribe = 101
)

//封包
func EnPack(protocol int, message []byte) []byte {
	return append(append(append([]byte(ConstHeader), IntToBytes(protocol)...), IntToBytes(len(message))...), message...)
}

var headerLength int

func init() {
	headerLength = len(ConstHeader)
}

//解包
func DePack(buffer []byte) (int, []byte) {
	length := len(buffer)

	var i, protocol int
	data := make([]byte, 0)
	for i = 0; i < length; i++ {
		if length < i+headerLength+ConstPLength+ConstMLength {
			return -1, data
		}
		if string(buffer[i:i+headerLength]) == ConstHeader {
			protocol = ByteToInt(buffer[i+headerLength : i+headerLength+ConstPLength])

			messageLength := ByteToInt(buffer[i+headerLength+ConstPLength : i+headerLength+ConstPLength+ConstMLength])
			if length < i+headerLength+ConstPLength+ConstMLength+messageLength {
				return 0, data
			}
			data = buffer[i+headerLength+ConstPLength+ConstMLength : i+headerLength+ConstPLength+ConstMLength+messageLength]
			break
		}
	}

	if i == length {
		return -2, data
	}

	return protocol, data
}

//字节转换成整形
func ByteToInt(n []byte) int {
	bytesBuffer := bytes.NewBuffer(n)
	var x int32
	binary.Read(bytesBuffer, binary.BigEndian, &x)

	return int(x)
}

//整数转换成字节
func IntToBytes(n int) []byte {
	x := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

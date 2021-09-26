package TCPNet

import (
	"encoding/binary"
	"errors"
	"github.com/restxx/GiRobot/stream"
	"github.com/restxx/GiRobot/utils"
	"log"
)

const (
	HeadLen = 3 // int(unsafe.Sizeof(PackHead{}))
)

// 如果包头类型变化这里uint16得改成对应类型 改
type PackHead struct {
	Len   uint16
	Flag1 uint8
}

type Packet struct {
	head *PackHead
	body *stream.Buffer
}

// 计算包函包头的长度的完整包长
// 当前head.Len表示包头+包体大小
func (pHead *PackHead) WholeLen() int {
	return int(pHead.Len)
}

func GetHead(pStream stream.IBuffIO, pHead *PackHead) (err error) {
	sz, err := pStream.Peek(int(HeadLen))
	if err != nil {
		return
	}
	_ = binary.Read(stream.NewBuffer(sz), utils.Endian, pHead)
	return
}

func CreatePacket() *Packet {
	return &Packet{
		head: &PackHead{},
		body: stream.NewBuffer([]byte{}),
	}
}

func (pak *Packet) AddData(src ...interface{}) {
	buffer := pak.body
	for _, v := range src {
		switch v.(type) {
		case string: // 目前没有使用过
			str := v.(string)
			_ = binary.Write(buffer, utils.Endian, uint16(len(str)))
			_, _ = buffer.WriteString(str)
		default:
			_ = binary.Write(buffer, utils.Endian, v)
		}
	}
}

// 如果包头类型变化这里uint16得改成对应类型 改
func (pak *Packet) ClosePacket(src ...interface{}) []byte {
	rBuffer := stream.NewBuffer([]byte{})

	// 写入PackHead
	head := pak.head
	head.Len = uint16(pak.body.Len()) + uint16(HeadLen)

	for i, v := range src {
		switch i {
		case 0:
			head.Flag1 = v.(uint8) // 加不加密
		}
	}
	_ = binary.Write(rBuffer, utils.Endian, head)

	_, _ = rBuffer.Write(pak.body.Bytes())
	return rBuffer.Bytes()
}

/*
 将msgBlock传过来的数据打包
 设计时原则上传入的Data字段应該是消息结构指針
 var buff []byte = msgBlocks.Data.(IMessage).ToBytes()
*/
func EnCode(msgBlocks *utils.MsgBlock, enKey []byte) (data []byte, err error) {

	err = nil
	pak := CreatePacket()
	pak.AddData(msgBlocks.Data)
	data = pak.ClosePacket(uint8(0))
	return
}

var Pool utils.MsgPool = utils.NewPool()

// 将收到的流式数据分解成包
func DeCode(pStream stream.IBuffIO, deKey []byte) (msgBlocks []*utils.MsgBlock, err error) {
	total, err := int(0), nil // 成功分解包个数

	for {
		var head PackHead
		if uErr := GetHead(pStream, &head); uErr != nil {
			if total == 0 {
				err = uErr
			}
			break
		}

		msg, err := pStream.Next(head.WholeLen())
		if err != nil {
			log.Println("Stream.UnreadLen:", pStream.Len())
			if total == 0 {
				err = errors.New("Packet Unready... ")
			}
			break
		}
		block := Pool.Get()

		if stream.BTUint8(msg[2:3]) == 0 {
			block.MID = stream.BTUint32(msg[3:7])
			block.SID = 0
			block.Write(msg[3:])
		}
		msgBlocks = append(msgBlocks, block)
		total += 1
	} // end for
	return
}

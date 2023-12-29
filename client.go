package main

import (
	"fmt"
	"net"
	"projectK/message"
	"projectK/pack"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
)

func main() {
	strIP := "localhost:12345"
	var conn net.Conn
	var err error

	//连接服务器
	if conn, err = net.Dial("tcp", strIP); err != nil {
		panic(err)
	}

	fmt.Println("connect", strIP, "success")
	defer conn.Close()

	//发送消息
	//sender := bufio.NewScanner(os.Stdin)
	for i := 0; i < 2; i++ {
		stSend := &message.User{
			Username: "wang " + strconv.Itoa(i),
			Passwd:   *proto.String("wang" + strconv.Itoa(i)),
		}
		//protobuf编码
		pData, err := proto.Marshal(stSend)
		if err != nil {
			panic(err)
		}
		data, err := pack.Encode(pData)
		if err == nil {
			conn.Write(data)
			fmt.Printf("%d send %d \n", i, uint16(len(pData)))
		}

		/*if sender.Text() == "stop" {
		   return
		}*/
	}
	for {
		time.Sleep(time.Millisecond * 10)
		/*stSend := &model.People{
		     Name: "lishang "+ strconv.Itoa(22222222),
		     Age:  *proto.Int(22222222),
		  }

		  //protobuf编码
		  pData, err := proto.Marshal(stSend)
		  if err != nil {
		     panic(err)
		  }
		  b := make([]byte,2,2)
		  binary.BigEndian.PutUint16(b,uint16(len(pData)))
		  //发送
		  conn.Write(append(b,pData...))*/
	}

}

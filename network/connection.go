package network

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"time"
)

const (
	CONN_DISCONNECT = 0
	CONN_CONNECTED  = 1
	CONN_STOP       = 2
)

type Connection struct {
	Conn         net.Conn
	Readtimeout  time.Duration
	Writetimeout time.Duration
	Reader       *bufio.Reader
	Writer       *bufio.Writer
	State        uint32
	s            *Server
	username     string
	password     string
	session_id   int64
}

func NewConn(Conn net.Conn, s *Server) *Connection {
	return &Connection{
		Conn:         Conn,
		Readtimeout:  time.Duration(10 * time.Second),
		Writetimeout: time.Duration(10 * time.Second),
		Reader:       bufio.NewReader(Conn),
		Writer:       bufio.NewWriter(Conn),
		State:        CONN_CONNECTED,
		s:            s,
		session_id:   s.node.Generate().Int64(),
	}
}

func (c *Connection) ReadCallBack() {
	for {
		buff := make([]byte, 1024)
		n, err := c.Conn.Read(buff)
		if err != nil && err != io.EOF {
			c.s.Logger.Println("读取用户消息错误")
			continue
		}
		if n == 0 {
			break
		}

		msg := string(buff[:n])
		fmt.Printf("<<<<<== server get client message:%s \n", msg)
		/*loginRequest := &message.LoginRequest{}
		loginResponse := &message.LoginResponse{}
		proto.Unmarshal(buf[:n], loginRequest)
		LoginProcess(loginRequest, loginResponse)

		// 序列化成二进制
		responsemsg, _ := proto.Marshal(loginResponse)
		fmt.Printf("responsemsg: %v\n", responsemsg)

		fmt.Printf("=>>>>>> server sent client message:%s \n", responsemsg)
		conn.Write([]byte(responsemsg))*/
	}
}

/*
	func LoginProcess(res *message.LoginRequest, req *message.LoginResponse) {
		fmt.Printf("//rpc LoginLogic//\n")
		fmt.Printf("res.GetName(): %s\n", res.GetName())
		fmt.Printf("res.GetPwd(): %s\n", res.GetPwd())

		if bSuccess := LoginCheck(res); bSuccess {
			req.Success = true
			req.Result = &message.ResultCode{
				Errcode: 123,
				Errmsg:  "yes",
			}
		} else {
			req.Success = false
			req.result = &message.ResultCode{
				Rrrcode: 1,
				Errmsg:  "no",
			}
		}
	}

	func (c *Connection) LoginCheck(msg *message.LoginRequest) bool {
		username := msg.Name
		pwd := msg.Pwd
		authRequest := AuthRequest{
			Session:  conn,
			username: username,
			password: pwd,
			result:   make(chan bool),
		}

		c.s.AuthChannel <- authRequest
		// 等待身份验证结果
		authResult := <-authRequest.result
		if !authResult {
			conn.Writer.Write([]byte("身份验证失败，请重新输入用户名和密码"))
			conn.Writer.Flush()
			return false
		}
		return true
	}
*/
func (c *Connection) WriteCallBack() {

}

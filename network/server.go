package network

import (
	"database/sql"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"projectK/conf"
	"projectK/message"
	"projectK/pack"
	"sync"
	"syscall"

	"github.com/bwmarrin/snowflake"
	"github.com/go-redis/redis/v8"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
)

var g_pack pack.DataPack

type Server struct {
	Name        string
	Ip          string
	Port        int
	Listener    *net.TCPListener
	QuitChan    chan struct{}
	Logger      *logrus.Logger
	Sessions    map[*Connection]bool //用户名：client
	connMutex   sync.Mutex
	Join        chan *Connection
	Leave       chan *Connection
	message     chan string
	loadMsg     chan struct{}
	node        *snowflake.Node
	database    *sql.DB
	redis       *redis.Client
	AuthChannel chan AuthRequest
	msgHandler  *pack.MsgHandler
}

func NewServer(logger *logrus.Logger, node *snowflake.Node, redis *redis.Client) *Server {
	config, err := conf.ReadConfig()
	if err != nil {
		fmt.Println("Reload config error", err)
		return nil
	}

	return &Server{
		Name:     config.Name,
		Ip:       config.GetListenAddress(),
		Port:     config.GetPort(),
		QuitChan: make(chan struct{}),
		Logger:   logger,
		Sessions: make(map[*Connection]bool),
		Join:     make(chan *Connection),
		Leave:    make(chan *Connection),
		message:  make(chan string),
		loadMsg:  make(chan struct{}),
		node:     node,
		//database: database,
		redis:       redis,
		AuthChannel: make(chan AuthRequest),
		msgHandler:  pack.NewMsgHandler(),
	}

}

func (s *Server) Start() {
	s.Serve()
}

func (s *Server) Stop() {

}

func (s *Server) AddSession(c *Connection) {
	s.connMutex.Lock()
	defer s.connMutex.Unlock()
	s.Sessions[c] = true
}

func (s *Server) RemoveSession(c *Connection) {
	s.connMutex.Lock()
	defer s.connMutex.Unlock()
	delete(s.Sessions, c)
}

func (s *Server) Accept() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			fmt.Println("[Server] Listener error")
			return
		}

		go s.HandleConn(conn)
	}
}

// 登录验证请求
type AuthRequest struct {
	Session  *Connection
	username string
	password string
	result   chan bool
}

func (s *Server) HandleConn(conn net.Conn) {
	c := NewConn(conn, s)
	s.handleSession(c)
}

func (s *Server) handleConnections() {
	for {
		select {
		case client := <-s.Join:
			fmt.Printf("%v 加入了游戏\n", client.username)
			s.AddSession(client)
			go s.broadcast(fmt.Sprintf("%s 加入了游戏 ,SessionID %d\n", client.username, client.session_id))
		case client := <-s.Leave:
			fmt.Printf("%v 离开了游戏", client.username)
			s.RemoveSession(client)
			go s.broadcast(fmt.Sprintf("%s 离开了游戏\n", client.username))
		case message := <-s.message:
			fmt.Printf("读取到 广播 消息 %s", message)
			go s.broadcast(message)
		case load := <-s.loadMsg:
			fmt.Println("加载数据:", load)
			sstream, _ := s.readMessage()
			fmt.Println("加载数据from redis:", sstream)
			for _, msg := range sstream {
				fmt.Println(msg)
			}
		case authRequest := <-s.AuthChannel:
			go s.authenticateUser(authRequest)
		}
	}
}

func (s *Server) authenticateUser(request AuthRequest) {
	storedHash, err := s.redis.Get(context.Background(), fmt.Sprintf("user:%s", request.username)).Result()
	if err == redis.Nil {
		err = s.createUser(request.username, request.password)
		if err != nil {
			fmt.Println("创建用户错误")
			request.result <- false
		}
		request.result <- true
		return
	} else if err != nil {
		fmt.Printf("获取用户信息错误：", err)
		request.result <- false
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(request.password))
	if err == nil {
		fmt.Println("用户验证 %s 成功", request.username)
		request.result <- true
	} else {
		fmt.Printf("用户%s验证密码失败\n", request.username)
		fmt.Printf("输入密码 %s \n", request.password)
		fmt.Printf("数据库密码 %s \n", storedHash)
		fmt.Printf("Error:%v", err)
		request.result <- false
	}
}

func (s *Server) createUser(username, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	fmt.Printf("创建用户%s成功,密码已加密\n", username)
	return s.redis.Set(context.Background(), fmt.Sprintf("user:%s", username), hashedPassword, 0).Err()
}

func (s *Server) handleSession(conn *Connection) {
	/*if inited := conn.LoginProcess(); !inited {
		return
	}*/
	conn.Writer.Write([]byte(fmt.Sprintf("欢迎 %v 您可以聊天了，输入'exit' 退出。\n", conn.username)))
	conn.Writer.Flush()
	s.Join <- conn
	defer func() {
		s.Leave <- conn
	}()

	for {
		data, err := pack.Decode(conn.Reader)
		if err != nil {
			if err != io.EOF {
				continue
			}
			fmt.Println("消息解析错误:", err)
			break
		}

		if len(data) > 0 {
			user := &message.User{}
			err = proto.Unmarshal(data, user)
			if err != nil {
				panic(err)
			}
			fmt.Printf("receive %s %+v \n", conn.Conn.RemoteAddr(), user)
		}

		/*	b, err := pack.Encode(msg)
			if err != nil {
				fmt.Println("proto err ", err)
				break
			}

			_, err = conn.Conn.Write(b)
			if err != nil {
				fmt.Println("send err ", err)
				break
			}
			message, err := conn.Reader.ReadString('\n')
			if err != nil {
				s.Logger.Println("读取客户端消息错误:", err)
				break
			}
			fmt.Println("读取消息 ：", message)

			if message == "exit\n" {
				break
			}

			if message == "load\n" {
				s.loadMsg <- struct{}{}
				continue
			}
			s.message <- fmt.Sprintf("[%s] %s: %s", time.Now().Format("2006-01-02 15:04:06"), conn.username, message)
			go s.storeMessage(conn.session_id, message)*/
	}

}

func (s *Server) storeMessage(session_id int64, message string) {
	fmt.Println("存储消息到redis：", message)
	err := s.redis.LPush(context.Background(), "chat_message", message).Err()
	if err != nil {
		fmt.Println("存储消息到redis错误", err)
	}
}

func (s *Server) readMessage() ([]string, error) {
	messages, err := s.redis.LRange(context.Background(), "chat_message", 0, 9).Result()
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (s *Server) getUserNameByID(user_id int64) string {
	for c := range s.Sessions {
		if c.session_id == user_id {
			return c.username
		}
	}
	return "None"
}

func InitDataBase() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "chat.db")
	if err != nil {
		fmt.Println("Open database error,chat.db is not exists")
		return nil, err
	}

	fmt.Println("Open chat.db success...")

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		message TEXT,
		timestamp DATETIME
	)`)

	fmt.Println("create table messages success...")

	if err != nil {
		return nil, err
	}

	return db, nil
}

func (s *Server) broadcast(message string) {
	/*msg := s.handleChatMessage([]byte(message))
	if msg == nil {
		fmt.Println("消息解析错误")
		return
	}*/
	for client := range s.Sessions {
		client.Writer.Write([]byte(message))
		client.Writer.Flush()
	}
}

func (s *Server) Serve() {
	var err error
	Addr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("[Server] ResolveTCPAddr error")
		return
	}

	s.Listener, err = net.ListenTCP("tcp4", Addr)
	if err != nil {
		fmt.Println("[Server] ListenTCP error")
		return
	}
	go s.handleConnections()

	s.Logger.Println("服务器已启动，监听地址:", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	go s.Accept()
	waitForShutDown()
}

func waitForShutDown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("收到关闭信号，正在优雅地关闭服务器...")
}

/*
func (s *Server) handleChatMessage(data []byte) *message.ChatMessage {
	s.msgHandler.GetHandler(message.MessageType_CHAT)(data)

	fmt.Printf("[%s]:%s\n", msg.Username, msg.Content)
	return msg
}*/

package pack

type Handle func(interface{}) bool

type MsgHandler struct {
	handlers map[uint32]Handle
}

func (m *MsgHandler) Register(msgId uint32, h Handle) {
	m.handlers[msgId] = h
}

func (m *MsgHandler) GetHandler(msgId uint32) Handle {
	if h, ok := m.handlers[msgId]; ok {
		return h
	}
	return nil
}

func NewMsgHandler() *MsgHandler {
	m := &MsgHandler{}
	return m
}

/*func ChatProcess(data interface{}) bool {
	d, ok := data.(*message.ChatMessageRequest)
	if ok {
		msg := &message.ChatMessageRequest{}
		err := proto.Unmarshal(d, msg)
		if err != nil {
			fmt.Println("解码消息失败：", err)
			return false
		}
	}
	return true
}*/

package pack

type Message struct {
	ID      uint32
	DataLen uint32
	Data    []byte
}

func NewMessage(ID uint32, DataLen uint32, Data []byte) *Message {
	return &Message{
		ID:      ID,
		DataLen: DataLen,
		Data:    Data,
	}
}

func (m *Message) GetID() uint32 {
	return m.ID
}

func (m *Message) GetDataLen() uint32 {
	return m.DataLen
}

func (m *Message) GetData() []byte {
	return m.Data
}

package pack

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	MESSAGE_MAX_LENGHT uint32 = 1024 * 1024
)

type DataPack struct {
}

func (dp *DataPack) Pack(m *Message) ([]byte, error) {
	dataBuffer := bytes.NewBuffer([]byte{})

	if err := binary.Write(dataBuffer, binary.BigEndian, m.GetID()); err != nil {
		return nil, err
	}

	fmt.Printf("pack message id  is %v,buffer: %v\n", m.GetID(), dataBuffer)
	if err := binary.Write(dataBuffer, binary.BigEndian, m.GetDataLen()); err != nil {
		return nil, err
	}

	if err := binary.Write(dataBuffer, binary.BigEndian, m.GetData()); err != nil {
		return nil, err
	}
	fmt.Println("pack message is", dataBuffer)
	return dataBuffer.Bytes(), nil
}

func (dp *DataPack) Unpack(data []byte) (*Message, error) {
	buffer := bytes.NewReader(data)
	m := &Message{}
	if err := binary.Read(buffer, binary.BigEndian, m.ID); err != nil {
		return nil, err
	}

	if err := binary.Read(buffer, binary.BigEndian, m.DataLen); err != nil {
		return nil, err
	}

	//
	if m.DataLen > MESSAGE_MAX_LENGHT {
		return nil, errors.New("data lenght range out of max lenght")
	}

	return m, nil
}

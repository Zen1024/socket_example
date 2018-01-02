package main

import (
	"encoding/json"
	"github.com/Zen1024/socket_proto"
)

type Message interface {
	GetMessageId() int32
}

const (
	ConnRequestMessageID           = int32(0)
	ConnResponseMessageID          = int32(1)
	ConnHeartBeatReqMessageID      = int32(2)
	ConnHeartBeatResponseMessageID = int32(3)
)

type ConnectRequest struct {
	UserId int `json:"user_id"`
}

func (m *ConnectRequest) GetMessageId() int32 {
	return ConnRequestMessageID
}

type ConnectResponse struct {
	HeartBeatInterval int `json:"heart_beat_interval"` //发送心跳包的间隔(s)
}

func (m *ConnectResponse) GetMessageId() int32 {
	return ConnResponseMessageID
}

type HeartBeatRequest struct {
	Timestamp int64 `json:"timestamp"`
}

func (m *HeartBeatRequest) GetMessageId() int32 {
	return ConnHeartBeatReqMessageID
}

type HeartBeatResponse struct {
	Timestamp int64 `json:"timestamp"` //原样返回时间戳
}

func (m *HeartBeatResponse) GetMessageId() int32 {
	return ConnHeartBeatResponseMessageID
}

func serializeMsg(m Message) ([]byte, error) {
	h := &proto.SocketHeader{}
	h.SetMessageID(m.GetMessageId())
	ctnt, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	p := proto.NewPacket(h, ctnt, nil)

	return p.Serialize(), nil
}

func marshalMsg(m Message) ([]byte, error) {
	ctnt, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return ctnt, nil
}

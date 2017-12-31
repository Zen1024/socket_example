package main

import (
	"encoding/json"
	"github.com/Zen1024/gosocket"
	"github.com/Zen1024/socket_proto"
	"time"
)

type Ctx struct{}

func (c *Ctx) OnConnect(*socket.Conn) bool {
	log.Print("OnConnect start......\n")
	defer log.Print("OnConnect end......\n")
	return true
}

func (c *Ctx) OnMessage(conn *socket.Conn, packet socket.ConnPacket) bool {
	log.Print("onMessage start......\n")
	defer log.Print("onMessage end......\n")

	reqPacket, ok := packet.(*proto.Packet)
	if !ok {
		log.Printf("invalid packet type\n")
		return false
	}
	if reqPacket.Handle != nil {
		reqPacket.Handle(conn, packet)
	}
	return true

}

func (c *Ctx) OnClose(*socket.Conn) {
	log.Print("OnClose start......\n")
	defer log.Print("OnClose end......\n")
}

func (c *Ctx) OnConnectRequest(conn *socket.Conn, packet socket.ConnPacket) {
	log.Print("OnConnRequest start.....\n")
	defer log.Print("OnConnRequest end......\n")
	p, _ := packet.(*proto.Packet)
	req := &ConnectRequest{}
	ctnt := p.GetContent()
	if err := json.Unmarshal(ctnt, req); err != nil {
		log.Printf("Error unmarshal:%s\n", err.Error())
		return
	}
	log.Printf("got req:%+v\n", req)
	respMsg := &ConnectResponse{
		HeartBeatInterval: 3,
	}
	bt, err := marshalMsg(respMsg)
	if err != nil {
		log.Printf("Error serialize msg:%s\n", err.Error())
	}

	h := &proto.SocketHeader{}
	pc := proto.NewPacket(h, bt, nil)

	if err := conn.AsyncWrite(pc, 3*time.Second); err != nil {
		log.Printf("Error write response:%s\n", err.Error())
	}

	return
}

func (c *Ctx) OnHeartBeatRequest(conn *socket.Conn, packet socket.ConnPacket) {
	log.Print("OnHeartBeatRequest start......\n")
	defer log.Print("OnHeartBeatRequest end......\n")
	p, _ := packet.(*proto.Packet)
	req := &HeartBeatRequest{}
	ctnt := p.GetContent()
	if err := json.Unmarshal(ctnt, req); err != nil {
		log.Printf("Error unmarshal:%s\n", err.Error())
		return
	}
	log.Printf("got req:%+v\n", req)
	respMsg := &HeartBeatResponse{
		Timestamp: req.Timestamp,
	}
	bt, err := marshalMsg(respMsg)
	if err != nil {
		log.Printf("Error serialize msg:%s\n", err.Error())
	}
	h := &proto.SocketHeader{}
	pc := proto.NewPacket(h, bt, nil)
	conn.AsyncWrite(pc, 3*time.Second)
	return
}

func main() {
	ctx := &Ctx{}
	mux := socket.NewMux()
	mux.Add(ConnRequestMessageID, "connect_request", ctx.OnConnectRequest)
	mux.Add(ConnHeartBeatReqMessageID, "heartBeat_request", ctx.OnHeartBeatRequest)
	p := &proto.Protocol{
		Mux: mux,
	}
	srv := socket.NewServer(":10000", 1000, ctx, p, 1000, 1000)
	go func() {
		time.Sleep(time.Second * 2)
		c := NewClient(":10000", 10)
		c.Run()
	}()
	srv.Serve()
}

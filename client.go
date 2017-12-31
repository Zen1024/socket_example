package main

import (
	"encoding/json"
	"fmt"
	"github.com/Zen1024/socket_proto"
	"net"
	"sync"
	"time"
)

type Client struct {
	ServerAddr string
	UserId     int
	Conn       *net.TCPConn
	Proto      *proto.Protocol
	sync.Mutex
}

func NewClient(addr string, id int) *Client {
	return &Client{
		ServerAddr: addr,
		UserId:     id,
		Proto:      &proto.Protocol{},
	}

}

func (c *Client) write(buf []byte) error {
	lb := len(buf)
	var total, n int
	var err error

	for total < lb && err == nil {
		n, err = c.Conn.Write(buf[total:])
		total += n
	}

	if total >= lb {
		err = nil
	}
	return err
}

func (c *Client) Write(data []byte, readTimeout time.Duration) (*proto.Packet, error) {
	c.Lock()
	defer c.Unlock()

	if err := c.write(data); err != nil {
		return nil, err
	}
	for {
		select {
		case <-time.After(readTimeout):
			return nil, fmt.Errorf("read timeout\n")
		default:
		}
		p, err := c.readPacket()
		if err != nil {
			continue
		}
		return p, nil
	}

}

func (c *Client) AsyncWrite(data []byte, timeout time.Duration) error {
	c.Lock()
	defer c.Unlock()
	for {
		select {
		case <-time.After(timeout):
			return fmt.Errorf("write timeoud\n")
		default:
		}
		if err := c.write(data); err != nil {
			return err
		}
		return nil
	}
}

func (c *Client) readPacket() (*proto.Packet, error) {
	re, err := c.Proto.ReadConnPacket(c.Conn)
	if err != nil {
		return nil, err
	}
	p, ok := re.(*proto.Packet)
	log.Printf("h:%+v\n", p.Header)
	if !ok {
		panic("invalid read packet return")
	}
	return p, nil
}

func (c *Client) connect() {
	c.Lock()
	defer c.Unlock()
	addr, err := net.ResolveTCPAddr("tcp4", c.ServerAddr)
	if err != nil {
		panic(err)
	}
	conn, err := net.DialTCP("tcp4", nil, addr)
	if err != nil {
		panic(err)
	}
	c.Conn = conn
	return
}

func (c *Client) doConnect() (*ConnectResponse, error) {
	log.Print("do Connect start......\n")
	defer log.Print("do Connect end......\n")
	req := &ConnectRequest{
		UserId: c.UserId,
	}
	bt, err := serializeMsg(req)
	if err != nil {
		return nil, err
	}
	re, err := c.Write(bt, time.Second)
	if err != nil {
		return nil, err
	}
	ctnt := re.GetContent()
	conn_resp := &ConnectResponse{}
	if err := json.Unmarshal(ctnt, conn_resp); err != nil {
		log.Printf("ctnt:%s\n", string(ctnt))
		log.Printf("Error unmarshal:%s\n", err.Error())
		return nil, err
	}
	return conn_resp, nil
}

func (c *Client) doHeartBeat() error {
	log.Print("do heartBeat start.....\n")
	defer log.Print("do heartBeat end......\n")
	now := time.Now().Unix()
	log.Printf("ts:%d\n", now)
	req := &HeartBeatRequest{
		Timestamp: now,
	}
	bt, err := serializeMsg(req)
	if err != nil {
		return err
	}

	re, err := c.Write(bt, time.Second)
	if err != nil {
		return err
	}
	ctnt := re.GetContent()
	resp := &HeartBeatResponse{}
	if err := json.Unmarshal(ctnt, resp); err != nil {
		log.Printf("Errur unmarshal:%s\n", err.Error())
		return err
	}
	log.Printf("got resp ts:%d\n", resp.Timestamp)
	return nil
}

func (c *Client) Run() {
	c.connect()
	re, err := c.doConnect()
	if err != nil {
		log.Printf("Error connect:%s\n", err.Error())
		return
	}
	log.Printf("connected,got heart beat interval:%d\n", re.HeartBeatInterval)
	dur := time.Second
	if re.HeartBeatInterval > 0 {
		dur = time.Second * time.Duration(re.HeartBeatInterval)
	}
	ticker := time.NewTicker(dur)
	for {
		select {
		case <-ticker.C:
			c.doHeartBeat()
		default:
		}
	}
}

package states

import (
	"fmt"
	"log"
	"server/internal/server"
	"server/pkg/packets"
)

//this will hold the connected state logic

type Connected struct {
	client server.ClientInterfacer
	logger *log.Logger
}

func (c *Connected) Name() string {
	return "connected"
}

func (c *Connected) SetClient(client server.ClientInterfacer) {
	c.client = client
	loggingPrefix := fmt.Sprintf("Client %d [%s]: ", client.Id(), c.Name())
	c.logger = log.New(log.Writer(), loggingPrefix, log.LstdFlags)
}

func (c *Connected) OnEnter() {
	// a newly connected client will want to know its if first
	c.client.SocketSend(packets.NewId(c.client.Id()))
	c.logger.Printf("Send Id to client")
}

func (c *Connected) HandleMessage(senderId uint64, message packets.Msg) {
	// we are setting c.id to client id recived so if in this func if senderId and c.id match we broadcast the message
	if senderId == c.client.Id() {
		// This message was sent by our own client, so broadcast it to everyone else
		c.client.Broadcast(message)
	} else {
		// Another client interfacer passed this onto us, or it was broadcast from the hub,
		// so forward it directly to our own client
		c.client.SocketSendAs(message, senderId)
	}
}

func (c *Connected) OnExit() {

}

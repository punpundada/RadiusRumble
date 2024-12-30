package states

import (
	"context"
	"errors"
	"fmt"
	"log"
	"server/internal/server"
	"server/internal/server/db"
	"server/pkg/packets"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

//this will hold the connected state logic

type Connected struct {
	client  server.ClientInterfacer
	logger  *log.Logger
	queries *db.Queries
	dbCtx   context.Context
}

func (c *Connected) Name() string {
	return "connected"
}

func (c *Connected) SetClient(client server.ClientInterfacer) {
	c.client = client
	c.queries = client.DbTx().Queries
	c.dbCtx = client.DbTx().Ctx

	loggingPrefix := fmt.Sprintf("Client %d [%s]: ", client.Id(), c.Name())
	c.logger = log.New(log.Writer(), loggingPrefix, log.LstdFlags)
}

func (c *Connected) OnEnter() {
	// a newly connected client will want to know its if first
	c.client.SocketSend(packets.NewId(c.client.Id()))
	c.logger.Printf("Send Id to client")
}

func (c *Connected) HandleMessage(senderId uint64, message packets.Msg) {
	// // we are setting c.id to client id recived so if in this func if senderId and c.id match we broadcast the message
	// if senderId == c.client.Id() {
	// 	// This message was sent by our own client, so broadcast it to everyone else
	// 	c.client.Broadcast(message)
	// } else {
	// 	// Another client interfacer passed this onto us, or it was broadcast from the hub,
	// 	// so forward it directly to our own client
	// 	c.client.SocketSendAs(message, senderId)
	// }

	switch message := message.(type) {
	case *packets.Packet_LoginRequest:
		c.handleLoginRequest(senderId, message)
	case *packets.Packet_RegisterRequest:
		c.handleRegisterRequest(senderId, message)
	}
}

func (c *Connected) OnExit() {

}

func (c *Connected) handleLoginRequest(senderId uint64, msg *packets.Packet_LoginRequest) {
	if c.client.Id() != senderId {
		c.logger.Printf("Received login request from diffrent client id (%d)", senderId)
		return
	}
	username := msg.LoginRequest.Username
	genereicFailMessage := packets.NewDenyResponse("Incorrect username or password")

	user, err := c.queries.GetUserByUsername(c.dbCtx, username)
	if err != nil {
		c.logger.Printf("Error getting user from username: %s", username)
		c.client.SocketSend(genereicFailMessage)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(msg.LoginRequest.Password))
	if err != nil {
		c.logger.Printf("Error getting user from username: %s", username)
		c.client.SocketSend(genereicFailMessage)
		return
	}
	c.logger.Printf("User %s logged in successfully", username)
	c.client.SocketSend(packets.NewOkResponse())
}

func (c *Connected) handleRegisterRequest(senderId uint64, msg *packets.Packet_RegisterRequest) {
	if senderId != c.client.Id() {
		c.logger.Printf("Register message received from another client:%d", senderId)
		return
	}
	userName := strings.ToLower(msg.RegisterRequest.Username)
	//validatin username
	err := validateUsername(userName)
	if err != nil {
		reason := fmt.Sprintf("Invalid username: %v", err)
		c.logger.Print(reason)
		c.client.SocketSend(packets.NewDenyResponse(reason))
		return
	}
	//check if user already exists
	_, err = c.queries.GetUserByUsername(c.dbCtx, userName)
	if err == nil {
		reason := fmt.Sprintf("User aleady exist: %s", userName)
		c.logger.Println(reason)
		c.client.SocketSend(packets.NewDenyResponse(reason))
	}
	genericFailMessage := packets.NewDenyResponse("Incorrect username or password")
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(msg.RegisterRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		c.logger.Printf("Failed to hash password: %s", userName)
		c.client.SocketSend(genericFailMessage)
		return
	}

	_, err = c.queries.CreateUser(c.dbCtx, db.CreateUserParams{
		Username:     userName,
		PasswordHash: string(passwordHash),
	})

	if err != nil {
		c.logger.Printf("Failed to create user %s: %v", userName, err)
		c.client.SocketSend(genericFailMessage)
		return
	}

	c.client.SocketSend(packets.NewOkResponse())

	c.logger.Printf("User %s registered successfully", userName)

}

func validateUsername(username string) error {
	if len(username) <= 0 {
		return errors.New("empty")
	}
	if len(username) > 20 {
		return errors.New("too long")
	}
	if username != strings.TrimSpace(username) {
		return errors.New("leading or trailing whitespace")
	}
	return nil
}

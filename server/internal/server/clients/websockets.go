package clients

import (
	"log"
	"net/http"
	"server/internal/server"
	"server/internal/server/states"
	"server/pkg/packets"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

// WebSocketClient should implement ClientInterfacer
type WebSocketClient struct {
	id       uint64
	conn     *websocket.Conn
	hub      *server.Hub
	sendChan chan *packets.Packet
	logger   *log.Logger
	state    server.ClientStateHandler
	dbTx     *server.DbTx
}

func NewWebSocketClient(
	hub *server.Hub,
	writer http.ResponseWriter,
	request *http.Request,
) (server.ClientInterfacer, error) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(_ *http.Request) bool { return true },
	}

	conn, err := upgrader.Upgrade(writer, request, nil)

	if err != nil {
		return nil, err
	}

	client := &WebSocketClient{
		hub:      hub,
		conn:     conn,
		sendChan: make(chan *packets.Packet, 256),
		logger:   log.New(log.Writer(), "Client unknown: ", log.LstdFlags),
		dbTx:     hub.NewDbTx(),
	}

	return client, nil
}

func (c *WebSocketClient) Initialize(id uint64) {
	c.id = id
	c.SetState(&states.Connected{})
}

func (c *WebSocketClient) Id() uint64 {
	return c.id
}

// ProcessMessage implements server.ClientInterfacer.
func (c *WebSocketClient) ProcessMessage(senderId uint64, message packets.Msg) {
	c.state.HandleMessage(senderId, message)
}

// SocketSend implements server.ClientInterfacer.
func (c *WebSocketClient) SocketSend(message packets.Msg) {
	c.SocketSendAs(message, c.id)
}

// SocketSendAs implements server.ClientInterfacer.
func (c *WebSocketClient) SocketSendAs(message packets.Msg, senderId uint64) {
	select {
	case c.sendChan <- &packets.Packet{SenderId: senderId, Msg: message}:
	default:
		c.logger.Printf("Client %d send channel full, dropping message: %T", c.id, message)
	}
}

// PassToPeer implements server.ClientInterfacer.
func (c *WebSocketClient) PassToPeer(message packets.Msg, peerId uint64) {
	if peer, exists := c.hub.Clients.Get(peerId); exists {
		peer.ProcessMessage(c.id, message)
	}
}

// Broadcast implements server.ClientInterfacer.
func (c *WebSocketClient) Broadcast(message packets.Msg) {
	c.hub.BroadcastChan <- &packets.Packet{SenderId: c.id, Msg: message}
}

// ReadPump implements server.ClientInterfacer.
func (c *WebSocketClient) ReadPump() {
	defer func() {
		c.logger.Println("Closing read pump")
		c.Close("read pump closed")
	}()

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Printf("error: %v", err)
			}
			break
		}

		packet := &packets.Packet{}
		err = proto.Unmarshal(data, packet)
		if err != nil {
			c.logger.Printf("error unmarshalling data: %v", err)
			continue
		}

		// To allow the client to lazily not set the sender ID, we'll assume they want to send it as themselves
		if packet.SenderId == 0 {
			packet.SenderId = c.id
		}

		c.ProcessMessage(packet.SenderId, packet.Msg)
	}

}

// WritePump implements server.ClientInterfacer.
func (c *WebSocketClient) WritePump() {

	defer func() {
		c.logger.Println("Closing write pump")
		c.Close("write pump closed")
	}()

	for packet := range c.sendChan {
		writer, err := c.conn.NextWriter(websocket.BinaryMessage)
		if err != nil {
			c.logger.Printf("error getting writer for %T packet, closing client: %v", packet.Msg, err)
			return
		}

		data, err := proto.Marshal(packet)
		if err != nil {
			c.logger.Printf("error marshalling %T packet, dropping: %v", packet.Msg, err)
			continue
		}

		_, writeErr := writer.Write(data)

		if writeErr != nil {
			c.logger.Printf("error writing %T packet: %v", packet.Msg, err)
			continue
		}
		// because protobuf messages are in binary we add new line to every message
		// so that the messages do not stick to gether
		writer.Write([]byte{'\n'})

		if closeErr := writer.Close(); closeErr != nil {
			c.logger.Printf("error closing writer, dropping %T packet: %v", packet.Msg, err)
			continue
		}
	}
}

// Close implements server.ClientInterfacer.
func (c *WebSocketClient) Close(reason string) {
	c.logger.Printf("Closing client connection because: %s", reason)

	c.hub.UnregisterChan <- c
	c.conn.Close()
	c.SetState(nil)
	if _, closed := <-c.sendChan; !closed {
		close(c.sendChan)
	}
}

// SetState implements server.ClientInterfacer.
func (c *WebSocketClient) SetState(newState server.ClientStateHandler) {
	prevStateName := "None"
	if c.state != nil {
		prevStateName = c.state.Name()
		c.state.OnExit()
	}

	newStateName := "None"
	if newState != nil {
		newStateName = newState.Name()
	}

	c.logger.Printf("Switching from state %s to %s", prevStateName, newStateName)

	c.state = newState

	if c.state != nil {
		c.state.SetClient(c)
		c.state.OnEnter()
	}
}

func (c *WebSocketClient) DbTx() *server.DbTx {
	return c.dbTx
}

func (c *WebSocketClient) SharedGameObjects() *server.SharedGameObjects {
	return c.hub.SharedGameObjects
}

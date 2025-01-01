package server

import (
	"context"
	"database/sql"
	_ "embed"
	"log"
	"net/http"
	"server/internal/server/db"
	"server/internal/server/objects"
	"server/pkg/packets"

	_ "modernc.org/sqlite"
)

//when we import package as _ these are blank imports
// used for only their side effects

// Embed the database schema to be used when creating the database tables
//
//go:embed db/config/schema.sql
var schemaGenSql string

type DbTx struct {
	Ctx     context.Context
	Queries *db.Queries
}

func (h *Hub) NewDbTx() *DbTx {
	return &DbTx{
		Ctx:     context.Background(),
		Queries: db.New(h.dbPool),
	}
}

// A structure for the connected client to interface with the hub
type ClientInterfacer interface {
	Id() uint64
	ProcessMessage(senderId uint64, message packets.Msg)

	// Sets the client's ID and anything else that needs to be initialized
	Initialize(id uint64)

	// Puts data from this client in the write pump
	SocketSend(message packets.Msg)

	// Puts data from another client in the write pump
	SocketSendAs(message packets.Msg, senderId uint64)

	// Forward message to another client for processing
	PassToPeer(message packets.Msg, peerId uint64)

	// Forward message to all other clients for processing
	Broadcast(message packets.Msg)

	// Pump data from the connected socket directly to the client
	ReadPump()

	// Pump data from the client directly to the connected socket
	WritePump()

	// Close the client's connections and cleanup
	Close(reason string)

	SetState(newState ClientStateHandler)

	// A reference to the database transaction context for this client
	DbTx() *DbTx

	SharedGameObjects() *SharedGameObjects
}

// The hub is the central point of communication between all connected clients
type Hub struct {
	Clients *objects.SharedCollection[ClientInterfacer]
	// Clients map[uint64]ClientInterfacer

	// Packets in this channel will be processed by all connected clients except the sender
	BroadcastChan chan *packets.Packet

	// Clients in this channel will be registered with the hub
	RegisterChan chan ClientInterfacer

	// Clients in this channel will be unregistered with the hub
	UnregisterChan chan ClientInterfacer
	//database connection pool
	dbPool *sql.DB

	SharedGameObjects *SharedGameObjects
}

func NewHub() *Hub {
	dbPool, err := sql.Open("sqlite", "db.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	return &Hub{
		Clients:        objects.NewSharedCollection[ClientInterfacer](10),
		BroadcastChan:  make(chan *packets.Packet),
		RegisterChan:   make(chan ClientInterfacer),
		UnregisterChan: make(chan ClientInterfacer),
		dbPool:         dbPool,
		SharedGameObjects: &SharedGameObjects{
			Players: objects.NewSharedCollection[*objects.Player](),
		},
	}
}

func (h *Hub) Run() {
	log.Println("Initializing database...")
	_, err := h.dbPool.ExecContext(context.Background(), schemaGenSql)
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case client := <-h.RegisterChan:
			client.Initialize(h.Clients.Add(client))
		case client := <-h.UnregisterChan:
			h.Clients.Remove(client.Id())

		case packet := <-h.BroadcastChan:
			h.Clients.ForEach(func(clientId uint64, client ClientInterfacer) {
				// we dont want to send packet to ourself
				if clientId != packet.SenderId {
					client.ProcessMessage(packet.SenderId, packet.Msg)
				}
			})
		}
	}
}

// Creates a client for the new connection and begins the concurrent read and write pumps
func (h *Hub) Serve(
	getNewClient func(*Hub, http.ResponseWriter, *http.Request) (ClientInterfacer, error),
	writer http.ResponseWriter,
	request *http.Request,
) {
	log.Println("New client connected from", request.RemoteAddr)
	client, err := getNewClient(h, writer, request)

	if err != nil {
		log.Printf("Error obtaining client for new connection: %v", err)
		return
	}

	h.RegisterChan <- client

	go client.WritePump()
	go client.ReadPump()
}

type ClientStateHandler interface {
	Name() string
	// Inject the client into the state handler
	SetClient(client ClientInterfacer)

	OnEnter()
	HandleMessage(senderId uint64, message packets.Msg)
	// cleanup state handler and perform any last actions
	OnExit()
}

type SharedGameObjects struct {
	// The id of player is id of client
	Players *objects.SharedCollection[*objects.Player]
}

package states

import (
	"fmt"
	"log"
	"math/rand"
	"server/internal/server"
	"server/internal/server/objects"
	"server/pkg/packets"
)

type InGame struct {
	client server.ClientInterfacer
	player *objects.Player
	logger *log.Logger
}

func (i *InGame) Name() string {
	return "InGame"
}

func (i *InGame) SetClient(client server.ClientInterfacer) {
	i.client = client
	loggingPrefix := fmt.Sprintf("Client %d [%s]: ", client.Id(), i.Name())
	i.logger = log.New(log.Writer(), loggingPrefix, log.LstdFlags)
}

func (i *InGame) OnEnter() {
	i.logger.Printf("Adding player %s to the shared collection", i.player.Name)
	i.client.SharedGameObjects().Players.Add(i.player, i.client.Id())

	i.player.X = rand.Float64() * 1000
	i.player.Y = rand.Float64() * 1000
	i.player.Speed = 150.0
	i.player.Direction = 20.0
	i.client.SocketSend(packets.NewPlayer(i.client.Id(), i.player))
}

func (i *InGame) HandleMessage(senderId uint64, message packets.Msg) {
	// switch message := message.(type) {
	// case *packets.Packet_Chat:
	// 	i.client.SocketSend(message)
	// }
}

func (i *InGame) OnExit() {
	i.client.SharedGameObjects().Players.Remove(i.client.Id())
}

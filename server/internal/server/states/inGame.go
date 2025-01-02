package states

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"server/internal/server"
	"server/internal/server/objects"
	"server/pkg/packets"
	"time"
)

type InGame struct {
	client                 server.ClientInterfacer
	player                 *objects.Player
	logger                 *log.Logger
	canclePlayerUpdateLoop context.CancelFunc
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
	switch message := message.(type) {
	case *packets.Packet_Player:
		i.handlePlayer(senderId, message)
	case *packets.Packet_PlayerDirection:
		i.handlePlayerDitrection(senderId, message)
	}
}

func (i *InGame) handlePlayerDitrection(senderId uint64, message *packets.Packet_PlayerDirection) {
	if senderId == i.client.Id() {
		i.player.Direction = message.PlayerDirection.Direction

		// if this is only first time receving plater direction message from client we will start update loop
		if i.canclePlayerUpdateLoop == nil {
			ctx, cancle := context.WithCancel(context.Background())
			i.canclePlayerUpdateLoop = cancle
			go i.playerUpdateLoop(ctx)
		}
	}
}

func (i *InGame) handlePlayer(senderId uint64, message *packets.Packet_Player) {
	if senderId == i.client.Id() {
		i.logger.Println("Received player message from our own client, ignoring")
		return
	}
	i.client.SocketSendAs(message, senderId)
}

func (i *InGame) OnExit() {
	i.client.SharedGameObjects().Players.Remove(i.client.Id())
	// if on exit of this state if update loop is not candled we will call the method ourself to cancle the loop and stop ticker
	if i.canclePlayerUpdateLoop != nil {
		i.canclePlayerUpdateLoop()
	}
}

// this function will be called by playerUpdateLoop
// which will call function every few miliseconds
func (i *InGame) SyncPlayer(delta float64) {
	newX := i.player.X + i.player.Speed*math.Cos(i.player.Direction)*delta
	newY := i.player.Y + i.player.Speed*math.Sin(i.player.Direction)*delta
	i.player.X = newX
	i.player.Y = newY

	updatedPlayerPacket := packets.NewPlayer(i.client.Id(), i.player)
	i.client.Broadcast(updatedPlayerPacket)
	go i.client.SocketSend(updatedPlayerPacket)

}

func (i *InGame) playerUpdateLoop(ctx context.Context) {
	const delta float64 = 0.05
	ticker := time.NewTicker(time.Duration(delta*1000) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			i.SyncPlayer(delta)
		case <-ctx.Done():
			return
		}
	}
}

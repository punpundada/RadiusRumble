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
	i.player.Radius = 20.0
	i.client.SocketSend(packets.NewPlayer(i.client.Id(), i.player))
	// send the spore info to client
	go func() {
		i.client.SharedGameObjects().Spores.ForEach(func(sporeId uint64, spore *objects.Spore) {
			time.Sleep(5 * time.Millisecond)
			i.client.SocketSend(packets.NewSpore(sporeId, spore))
		})
	}()
}

func (i *InGame) HandleMessage(senderId uint64, message packets.Msg) {
	switch message := message.(type) {
	case *packets.Packet_Player:
		i.handlePlayer(senderId, message)
	case *packets.Packet_PlayerDirection:
		i.handlePlayerDitrection(senderId, message)
	case *packets.Packet_Chat:
		i.handleChat(senderId, message)
	case *packets.Packet_SporeConsumed:
		i.handleSporeConsumed(senderId, message)
	}
}

func (i *InGame) handleSporeConsumed(senderId uint64, message *packets.Packet_SporeConsumed) {
	// i.logger.Printf("Spore Consumed %d consumed by client %d", message.SporeConsumed.SporeId, senderId)
	if senderId != i.client.Id() {
		i.client.SocketSendAs(message, senderId)
		return
	}
	// If the spore was supposedly consumed by our own player, we need to verify the plausibility of the event
	errMsg := "Could not verify spore consumption: "

	spore, err := i.getSpore(message.SporeConsumed.SporeId)

	if err != nil {
		i.logger.Println(errMsg + err.Error())
		return
	}
	err = i.validatePlaterCloseToObject(spore.X, spore.Y, spore.Radius, 10)

	if err != nil {
		i.logger.Println(errMsg + err.Error())
		return
	}
	// If we made it this far, the spore consumption is valid, so grow the player, remove the spore, and broadcast the event
	sporeMass := radToMass(spore.Radius)
	i.player.Radius = i.nextRadius(sporeMass)
	go i.client.SharedGameObjects().Spores.Remove(message.SporeConsumed.SporeId)
	i.client.Broadcast(message)

}

func (i *InGame) handleChat(senderId uint64, message *packets.Packet_Chat) {
	if senderId == i.client.Id() {
		i.client.Broadcast(message)
	} else {
		i.client.SocketSendAs(message, senderId)
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

func (i *InGame) getSpore(sporeId uint64) (*objects.Spore, error) {
	spore, exist := i.client.SharedGameObjects().Spores.Get(sporeId)
	if !exist {
		return nil, fmt.Errorf("spore with id %d does not exits", sporeId)
	}
	return spore, nil
}

func (i *InGame) validatePlaterCloseToObject(objX float64, objY float64, radius float64, buffer float64) error {
	realDx := i.player.X - objX
	readlDy := i.player.Y - objY
	realDistSqr := realDx*realDx + readlDy*readlDy

	thresholedDist := i.player.Radius + buffer + radius
	thresholedDistSqrt := thresholedDist * thresholedDist
	if realDistSqr > thresholedDistSqrt {
		return fmt.Errorf(
			"Player is too far away from spore (distSq: %f, thresholedSq: %f)", realDistSqr, thresholedDist,
		)
	}
	return nil
}

func radToMass(radius float64) float64 {
	return math.Pi * radius * radius
}

func massToRad(mass float64) float64 {
	return math.Sqrt(mass / math.Pi)
}

func (i *InGame) nextRadius(massDiff float64) float64 {
	oldMass := radToMass(i.player.Radius)
	newMass := oldMass + massDiff
	return massToRad(newMass)
}

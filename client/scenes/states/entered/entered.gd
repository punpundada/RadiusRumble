extends Node2D

const packets := preload("res://pacets.gd")
const ws_url:String = "ws://localhost:8080/ws"
@onready var log: Log = $UI/Log

func _ready() -> void:
	WS.connected_to_server.connect(on_ws_connected)
	WS.packet_received.connect(on_ws_packet_received)
	WS.connection_closed.connect(on_ws_connection_closed)
	WS.connect_to_url(ws_url)
	log.info("Connecting to server...")
	
func on_ws_connected():
	log.success("connected successfully")

func on_ws_connection_closed():
	log.info("Connection closed")

func on_ws_packet_received(packet:packets.Packet)->void:
	var sender_id := packet.get_sender_id()
	if packet.has_id():
		_handle_id_msg(sender_id, packet.get_id())
#		chat will be handled by InGame state which we are setting once we get the client id in
#		_handle_id_msg function
	
	

func _handle_id_msg(sender_id: int, id_msg: packets.IdMessage) -> void:
	GameManager.client_id = id_msg.get_id()
	log.info("Received client ID: %d" % GameManager.client_id)
	GameManager.set_state(GameManager.State.INGAME)

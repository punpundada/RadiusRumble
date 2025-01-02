extends Node

@onready var log: Log = $UI/Log
@onready var line_edit: LineEdit = $UI/LineEdit
@onready var world: Node2D = $World
const packets = preload("res://packets.gd")
const Actor = preload("res://objects/actor/actor.gd")
var players := Dictionary()


func _ready() -> void:
	WS.connection_closed.connect(on_ws_connection_closed)
	WS.packet_received.connect(on_ws_packet_received)
	
	line_edit.text_submitted.connect(on_line_edit_text_submit)
	#log.info("You are -> %d" % GameManager.client_id)
	
	
func on_ws_connection_closed():
	log.info("Connection closed")

func on_ws_packet_received(packet:packets.Packet)->void:
	var sender_id := packet.get_sender_id()
	if packet.has_chat():
		_handle_chat_msg(sender_id, packet.get_chat())
	elif packet.has_player():
		_handle_player_msg(packet.get_sender_id(),packet.get_player())


func _handle_player_msg(senderId:int,player:packets.PlayerMessage)->void:
	var actor_id := player.get_id()
	var actor_name := player.get_name()
	var x := player.get_x()
	var y := player.get_y()
	var radius := player.get_radius()
	var speed := player.get_speed()
	
	var is_player := actor_id == GameManager.client_id
	if not players.has(actor_id):
#		this is a new player, so we need to create a new player 
		var actor := Actor.instantiate(actor_id,actor_name,x,y,radius,speed,is_player)
		world.add_child(actor)
		players[actor_id]=actor
	else:
		var actor = players[actor_id] as Actor
		actor.position.x = x
		actor.position.y = y
		
		
func _handle_chat_msg(sender_id: int, chat_msg: packets.ChatMessage) -> void:
	log.chat("Client %d" % sender_id, chat_msg.get_msg())
	
func on_line_edit_text_submit(text:String)->void:
	var packet := packets.Packet.new()
	var chat_msg := packet.new_chat()
	chat_msg.set_msg(text)
	
	var err := WS.send(packet)
	if err:
		log.error("Error sending chat message")
	else:
		log.chat("You", text)
	line_edit.text = ""

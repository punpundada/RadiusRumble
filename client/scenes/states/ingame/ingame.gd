extends Node

@onready var log: Log = $UI/Log
@onready var line_edit: LineEdit = $UI/LineEdit
@onready var world: Node2D = $World
const Spore = preload("res://objects/spore/spore.gd")
const packets = preload("res://packets.gd")
const Actor = preload("res://objects/actor/actor.gd")
var players := Dictionary()
var spores := Dictionary()

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
	elif packet.has_spore():
		_handle_spore_msg(sender_id,packet.get_spore())

func _handle_spore_msg(sender_id:int,packet:packets.SporeMessage):
	var spore_id := packet.get_id()
	var x := packet.get_x()
	var y := packet.get_y()
	var rad := packet.get_radius()
	
	if not spores.has(sender_id):
		var spore := Spore.instanciate(spore_id,x,y,rad)
		world.add_child(spore)
		spores[spore_id]=spore
		
func _handle_player_msg(senderId:int,player_packet:packets.PlayerMessage)->void:
	var actor_id := player_packet.get_id()
	var actor_name := player_packet.get_name()
	var x := player_packet.get_x()
	var y := player_packet.get_y()
	var radius := player_packet.get_radius()
	var speed := player_packet.get_speed()
	
	var is_player := actor_id == GameManager.client_id
	if not players.has(actor_id):
#		this is a new player, so we need to create a new player 
		add_actor(actor_id,actor_name,x,y,radius,speed,is_player)
	else:
		update_actor(actor_id,x,y,radius,speed,player_packet)

func add_actor(actor_id:int,actor_name:String,x:float,y:float,radius:float,speed:float,is_player:bool):
	var actor := Actor.instantiate(actor_id,actor_name,x,y,radius,speed,is_player)
	world.add_child(actor)
	players[actor_id]=actor
	if is_player:
		actor.area_entered.connect(_on_player_area_entered)
		
func update_actor(actor_id:int,x:float,y:float,radius:float,speed:float,player_packet:packets.PlayerMessage):
	var actor = players[actor_id] as Actor
	actor.position.x = x
	actor.position.y = y
	var direction := player_packet.get_direction()
	actor.velocity = speed * Vector2.from_angle(direction)

func _on_player_area_entered(area:Area2D):
	if area is Spore:
		consume_spore(area)

func consume_spore(spore:Spore):
	var packet := packets.Packet.new()
	var consume_packet := packet.new_spore_consumed()
	consume_packet.set_spore_id(spore.spore_id)
	WS.send(packet)
	remove_spore(spore)

func remove_spore(spore:Spore):
	spores.erase(spore.spore_id)
	spore.queue_free()
	

func _handle_chat_msg(sender_id: int, chat_msg: packets.ChatMessage) -> void:
	if players.has(sender_id):
		var actor:Actor = players[sender_id]
		log.chat(actor.actor_name, chat_msg.get_msg())
		
	
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

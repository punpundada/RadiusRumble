extends Node2D

const packets := preload("res://pacets.gd")
const ws_url:String = "ws://localhost:8080/ws"
@onready var log: Log = $Log
var client_id:int
@onready var line_edit: LineEdit = $LineEdit


func _ready() -> void:
	WS.connected_to_server.connect(on_ws_connected)
	WS.connection_closed.connect(on_ws_connection_closed)
	WS.packet_received.connect(on_ws_packet_received)
	
	log.info("Connecting to server...")
	WS.connect_to_url(ws_url)
	line_edit.text_submitted.connect(on_line_edit_text_submit)

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
	
func on_ws_connected():
	log.success("connected successfully")
	#var packet := packets.Packet.new()
	#var chat_msg = packet.new_chat()
	#chat_msg.set_msg("Hello Golang!")
	#
	#
	#var err := WS.send(packet)
	#if err != OK:
		#log.error("Error sending packet")
	#else:
		#log.success("Packet send")

func on_ws_connection_closed():
	log.error("Connection closed")

func on_ws_packet_received(packet:packets.Packet)->void:
	#log.info("Packet received: "+packet.to_string())
	var sender_id := packet.get_sender_id()
	if packet.has_id():
		_handle_id_msg(sender_id, packet.get_id())
	elif packet.has_chat():
		_handle_chat_msg(sender_id, packet.get_chat())
	
	

func _handle_id_msg(sender_id: int, id_msg: packets.IdMessage) -> void:
	var client_id = id_msg.get_id()
	log.info("Received client ID: %d" % client_id)
	
	
func _handle_chat_msg(sender_id: int, chat_msg: packets.ChatMessage) -> void:
	log.chat("Client %d" % sender_id, chat_msg.get_msg())

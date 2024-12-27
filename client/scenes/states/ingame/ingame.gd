extends Node2D

const packets = preload("res://pacets.gd")
@onready var log: Log = $UI/Log
@onready var line_edit: LineEdit = $UI/LineEdit


func _ready() -> void:
	WS.connection_closed.connect(on_ws_connection_closed)
	WS.packet_received.connect(on_ws_packet_received)
		
	line_edit.text_submitted.connect(on_line_edit_text_submit)
	log.info("You are -> %d" % GameManager.client_id)
	
	
func on_ws_connection_closed():
	log.info("Connection closed")

func on_ws_packet_received(packet:packets.Packet)->void:
	var sender_id := packet.get_sender_id()
	if packet.has_chat():
		_handle_chat_msg(sender_id, packet.get_chat())

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

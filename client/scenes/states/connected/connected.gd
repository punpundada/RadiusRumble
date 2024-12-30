extends Node
const packets = preload("res://packets.gd")

@onready var user_name_line_edit: LineEdit = $UI/VBoxContainer/UserNameLineEdit
@onready var password_line_edit: LineEdit = $UI/VBoxContainer/PasswordLineEdit
@onready var login: Button = $UI/VBoxContainer/HBoxContainer/Login
@onready var register: Button = $UI/VBoxContainer/HBoxContainer/Register
@onready var log: Log = $UI/VBoxContainer/Log

var action_on_ok_received: Callable

func _ready() -> void:
	WS.packet_received.connect(on_packet_received)
	WS.connection_closed.connect(on_connection_closed)
	
func on_packet_received(packet: packets.Packet):
	var senderId:int = packet.get_sender_id()
	if packet.has_deny_response():
		var deny_response := packet.get_deny_response()
		log.error(deny_response.get_reason())
	elif packet.has_ok_response():
		action_on_ok_received.call()

func on_connection_closed()->void:
	pass


func _on_login_pressed() -> void:
	var packet := packets.Packet.new()
	var login_request_message := packet.new_login_request()
	login_request_message.set_username(user_name_line_edit.text)
	login_request_message.set_password(password_line_edit.text)
	WS.send(packet)
	action_on_ok_received = func(): GameManager.set_state(GameManager.State.INGAME)
	
	


func _on_register_pressed() -> void:
	var packet := packets.Packet.new()
	var register_req_message = packet.new_register_request()
	register_req_message.set_username(user_name_line_edit.text)
	register_req_message.set_password(password_line_edit.text)
	WS.send(packet)
	action_on_ok_received = func():log.success("Registered Successfully")
	
	

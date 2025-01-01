extends Area2D

const packets = preload("res://packets.gd")
const Scene:PackedScene = preload("res://objects/actor/Actor.tscn")
const Actor = preload("res://objects/actor/actor.gd")

var actor_id:int
var actor_name:String
var start_x:float
var start_y:float
var start_rad:float	
var speed:float
var is_player:bool

var velocity:Vector2
var radius:float

@onready var collision_shape: CircleShape2D = $CollisionShape2D.shape
@onready var nameplate: Label = $Label
@onready var camera: Camera2D = $Camera2D

static func instantiate(
	actor_id:int,
	actor_name:String,
	x:float,
	y:float,
	radius:float,
	speed:float,
	is_player:bool,
	)->Actor:
	
	var actor = Scene.instantiate()
	actor.actor_id 
	actor.actor_name=actor_name
	actor.start_x=x
	actor.start_y=y
	actor.start_rad=radius
	actor.speed=speed
	actor.is_player=is_player
	
	return actor

func _ready() -> void:
	position.x = start_x
	position.y=start_y
	velocity=Vector2.RIGHT*speed
	radius =start_rad
	
	collision_shape.radius=radius
	nameplate.text=actor_name
	
	
func _physics_process(delta: float) -> void:
	position += velocity * delta
	if not is_player:
		return
	var mouse_pos = get_global_mouse_position()
	var input_vec := position.direction_to(mouse_pos).normalized()
	if abs(velocity.angle_to(input_vec)) > TAU/15: #12 deg
		velocity = input_vec *speed
		var packet = packets.Packet.new()
		var player_direction_msg = packet.new_player_direction()
		player_direction_msg.set_direction(velocity.angle())
		WS.send(packet)
	
#draw will be called once after ready is called
#if we want to call draw again we must call queue_redraw()
func _draw() -> void:
	draw_circle(Vector2.ZERO, collision_shape.radius, Color.DARK_ORCHID)

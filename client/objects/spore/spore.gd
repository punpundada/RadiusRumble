extends Area2D
const Spore = preload("res://objects/spore/spore.gd")
const Scene = preload("res://objects/spore/spore.tscn")
@onready var collision_shape = $CollisionShape2D.shape as CircleShape2D


var spore_id:int
var x:float
var y:float
var radius:float
var color:Color

static func instanciate(spore_id:int, x:float,y:float,redius:float) -> Spore:
	var spore_scene = Scene.instantiate() as Spore
	spore_scene.x=x
	spore_scene.y=y
	spore_scene.radius = redius
	spore_scene.spore_id = spore_id
	return spore_scene

func _ready() -> void:
	position.x = x
	position.y=y
	collision_shape.radius = radius
	color = Color.from_hsv(randf(),1,1,1)
	
	
func _draw() -> void:
	draw_circle(Vector2.ZERO,radius,color)

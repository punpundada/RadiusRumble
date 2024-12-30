extends Node


enum State {
	ENTERED,
	INGAME,
	CONNECTED,
}

var states_scenes: Dictionary = {
	State.ENTERED : "res://scenes/states/entered/entered.tscn",
	State.INGAME:"res://scenes/states/ingame/ingame.tscn",
	State.CONNECTED:"res://scenes/states/connected/connected.tscn",
}

var client_id: int
var current_scene_root: Node

func set_state(state: State) -> void:
	if current_scene_root != null:
		current_scene_root.queue_free()

	var scene: PackedScene = load(states_scenes[state])
	current_scene_root = scene.instantiate()
	add_child(current_scene_root)

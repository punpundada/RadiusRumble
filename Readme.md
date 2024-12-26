
# RadiusRumble

RadiusRumble is an exciting multiplayer game built with the Godot Engine. In this game, you play as a colored ball navigating a dynamic world filled with smaller balls and other players. Your goal is to consume smaller balls to grow in size and compete with other players. As you consume others, your ball's radius expands, making you an even greater threat—but also a bigger target!


## Features

- Dynamic Gameplay: Absorb smaller balls to grow your radius and dominate the field.
- Player Competition: Challenge other players in a battle of size and strategy.
- Scaling Difficulty: As your ball gets larger, your agility decreases, adding strategic depth.
- Cross platform because of Godot
- Godot-Powered Graphics: Smooth, vibrant visuals and physics-driven mechanics.


## Tech Stack

**Client:** (Game Engine) Godot Engine

**Server:**  Go (Golang) for handling server-side logic and multiplayer interactions


## Run Locally

Clone the project

```bash
  git clone https://github.com/punpundada/RadiusRumble.git
```

Go to the project directory

```bash
  cd RadiusRumble
```
### Install Dependencies

Go to the server directory
```bash
  cd server
```
Install Dependencies
```bash
  go mod tidy
```

### Godot 
No additional dependencies are required for Godot projects. Ensure you have Godot installed (V4.3 or above) on your system.

## Run go Backend

Navigate to server directory if not already

```bash
cd server
```

Run server
```bash
go run cmd/main.go
```

## Run Godot Client

Navigate to client directory
```bash
cd client
```

Open Godot project
```bash
godot .
```
Run the game from the Godot editor or export it to your desired platform.
## Contributing

Contributions are always welcome!



package ctypes

import (
	"errors"
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

type (
	PlayerFrameState int
	PlayerColour     int
	playerDirection  bool
)

const (
	frameUpdateTimeMillis = 50
	deltaX                = 2
)

// player frame states.
const (
	PlayerStanding PlayerFrameState = iota
	PlayerRunningStage0
	PlayerRunningStage1
	PlayerRunningStage2
	PlayerRunningStage4
	PlayerJumping
	PlayerCrouching
	PlayerMinState = PlayerStanding
	PlayerMaxState = PlayerCrouching
)

var playerRunningDecrease = false

func (frame PlayerFrameState) IsRunning() bool {
	return frame >= PlayerRunningStage0 && frame <= PlayerRunningStage4
}

func (frame *PlayerFrameState) Standing() {
	playerRunningDecrease = false
	*frame = PlayerStanding
}

func (frame *PlayerFrameState) Crouching() {
	playerRunningDecrease = false
	*frame = PlayerCrouching
}

func (frame *PlayerFrameState) Jumping() {
	playerRunningDecrease = false
	*frame = PlayerJumping
}

func (frame *PlayerFrameState) Running(lastUpdate *time.Time, changedDirection bool) {
	if time.Since(*lastUpdate).Milliseconds() < frameUpdateTimeMillis && !changedDirection {
		return
	}

	if *frame < PlayerRunningStage0 || *frame > PlayerRunningStage4 {
		*frame = PlayerRunningStage0
		return
	}

	if playerRunningDecrease {
		if *frame > PlayerRunningStage0 {
			*frame--
		} else {
			playerRunningDecrease = false
		}
	} else {
		if *frame < PlayerRunningStage4 {
			*frame++
		} else {
			playerRunningDecrease = true
		}
	}

	*lastUpdate = time.Now()
}

// player colours/types.
const (
	PlayerBlue PlayerColour = iota
	PlayerGreen
	PlayerPurple
	PlayerOrange
	PlayerUnknown
	PlayerMinColour = PlayerBlue
	PlayerMaxColour = PlayerOrange
)

func PlayerColourFromInt(i int) PlayerColour {
	var colour PlayerColour

	switch i {
	case 0:
		colour = PlayerBlue
	case 1:
		colour = PlayerGreen
	case 2:
		colour = PlayerPurple
	case 3:
		colour = PlayerOrange
	default:
		colour = PlayerUnknown
	}

	return colour
}

func (colour *PlayerColour) String() string {
	var str string

	switch *colour {
	case PlayerBlue:
		str = "Blue"
	case PlayerGreen:
		str = "Green"
	case PlayerPurple:
		str = "Purple"
	case PlayerOrange:
		str = "Orange"
	default:
		str = "Unknown"
	}

	return str
}

const (
	playerDirectionLeft  playerDirection = false
	playerDirectionRight playerDirection = true
)

func (direction *playerDirection) String() string {
	var str string

	switch *direction {
	case playerDirectionLeft:
		str = "left"
	case playerDirectionRight:
		str = "right"
	default:
		str = "unknown"
	}

	return str
}

func (direction playerDirection) MarshalJSON() ([]byte, error) {
	var str string

	switch direction {
	case playerDirectionLeft, playerDirectionRight:
		str = fmt.Sprintf("%q", direction.String())
	default:
		return nil, fmt.Errorf("playerDirection marshal error: invalid input %s", direction.String())
	}

	return []byte(str), nil
}

func (direction *playerDirection) UnmarshalJSON(data []byte) error {
	str := string(data)

	switch str {
	case "0", "false", "\"left\"":
		*direction = playerDirectionLeft
	case "1", "true", "\"right\"":
		*direction = playerDirectionRight
	default:
		return fmt.Errorf("playerDirection unmarshal error: invalid input %s", str)
	}

	return nil
}

type Player struct {
	Position             Position        `json:"pos,omitempty"`
	PlayerSpriteIndex    PlayerColour    `json:"sprite_index,omitempty"`
	Facing               playerDirection `json:"facing,omitempty"`
	lastFrameUpdate      time.Time
	frames               []*ebiten.Image
	playerAnimationFrame PlayerFrameState
	geoMatrix            ebiten.GeoM
	spritesheet          *Spritesheet
}

func getFrames(spritesheet *Spritesheet, spriteColour PlayerColour) ([]*ebiten.Image, error) {
	return spritesheet.GetPlayer(spriteColour)
}

/*
NewPlayer creates a new Player struct.

spritesheet is a non-optional parameter, and a nil value being passed in will result in
an error.

spriteIndex refers to the amount player sprite types available in the tilemap at
resources/images/tilemap_transparent_packed.png.
*/
func NewPlayer(spriteColour PlayerColour, spritesheet *Spritesheet, position Position) (*Player, error) {
	if spritesheet == nil {
		return nil, errors.New("spritesheet must not be nil")
	}

	frames, err := getFrames(spritesheet, spriteColour)
	if err != nil {
		return nil, err
	}

	matrix := ebiten.GeoM{}
	matrix.Translate(position.X, position.Y)

	return &Player{
		Position:          position,
		PlayerSpriteIndex: spriteColour,
		Facing:            playerDirectionRight,
		frames:            frames,
		geoMatrix:         matrix,
		spritesheet:       spritesheet,
	}, nil
}

func (p *Player) Update() {
	var didMove bool

	switch {
	case ebiten.IsKeyPressed(ebiten.KeyA):
		p.Position.AffectX(-deltaX)
		p.geoMatrix.Reset()

		p.geoMatrix.Scale(-1, 1)
		p.geoMatrix.Translate(SpriteSize, 0)

		if p.Facing == playerDirectionRight {
			p.Facing = playerDirectionLeft
			p.playerAnimationFrame.Running(&p.lastFrameUpdate, true)
		} else {
			p.playerAnimationFrame.Running(&p.lastFrameUpdate, false)
		}

		didMove = true
	case ebiten.IsKeyPressed(ebiten.KeyD):
		p.Position.AffectX(deltaX)
		p.geoMatrix.Reset()

		if p.Facing == playerDirectionLeft {
			p.Facing = playerDirectionRight
			p.playerAnimationFrame.Running(&p.lastFrameUpdate, true)
		} else {
			p.playerAnimationFrame.Running(&p.lastFrameUpdate, false)
		}

		didMove = true
	case ebiten.IsKeyPressed(ebiten.KeyS):
		p.playerAnimationFrame.Crouching()
	default:
		p.playerAnimationFrame.Standing()
	}

	if ebiten.IsKeyPressed(ebiten.KeyW) {
		p.Position.AffectY(deltaX)
		p.geoMatrix.Reset()

		p.playerAnimationFrame.Jumping()

		didMove = true
	}

	if didMove {
		p.geoMatrix.Translate(p.Position.X, p.Position.Y)
	}
}

func (p *Player) TickPhysics() {
	p.Position.AffectY(-4)
	p.geoMatrix.Reset()

	p.playerAnimationFrame.Jumping()
	p.geoMatrix.Translate(p.Position.X, p.Position.Y)
}

func (p *Player) InitFrames(spritesheet *Spritesheet) {
	if p.frames == nil {
		frames, err := getFrames(spritesheet, p.PlayerSpriteIndex)
		if err != nil {
			return
		}

		p.frames = frames
		p.spritesheet = spritesheet
	}
}

func (p *Player) RemoteUpdatePosition() {
	p.geoMatrix.Reset()

	p.geoMatrix.Scale(-1, 1)
	p.geoMatrix.Translate(SpriteSize, 0)
	p.geoMatrix.Translate(p.Position.X, p.Position.Y)
}

func (p *Player) Draw(screen *ebiten.Image) {
	if p.frames != nil {
		op := &ebiten.DrawImageOptions{GeoM: p.geoMatrix}
		screen.DrawImage(p.frames[p.playerAnimationFrame], op)
	}
}

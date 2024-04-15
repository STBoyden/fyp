package game

import (
	"fmt"
	"image"
	"io"
	"os"
	"strings"

	"fyp/common/ctypes"
	"fyp/common/ctypes/tiles"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	maxMapWidth  = 40
	maxMapHeight = 30
)

type Map struct {
	positions map[tiles.Types][]ctypes.Position
}

func LoadMapFromFile(path string) (*Map, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not load map from file \"%s\": %w", path, err)
	}

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("could not load map from file \"%s\": %w", path, err)
	}

	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("could not load map from file \"%s\": not a file", path)
	}

	var m Map
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("could not load map from file \"%s\": %w", path, err)
	}

	m.positions = make(map[tiles.Types][]ctypes.Position)
	m.fillPositions(string(content))

	return &m, nil
}

func (m *Map) fillPositions(content string) {
	for line, tileString := range strings.Split(content, "\n") {
		if line == maxMapHeight {
			break
		}

		for i, tileRune := range tileString {
			if i != 0 && i%maxMapWidth == 0 {
				line++
				i = 0
			}

			x := float64(i) * ctypes.SpriteSizeF
			y := float64(line) * ctypes.SpriteSizeF
			position := ctypes.NewPosition(x, y)

			for _, t := range tiles.Typeses.All() {
				if t.Symbol != string(tileRune) {
					continue
				}

				if positions, ok := m.positions[t]; ok {
					m.positions[t] = append(positions, position)
				} else {
					m.positions[t] = []ctypes.Position{position}
				}

				break
			}
		}
	}
}

func (m *Map) GetSpawnPoint() (x, y float64) {
	if positions, ok := m.positions[tiles.Typeses.SPAWNPOINT_TILE]; ok {
		pos := positions[0]

		return pos.X, pos.Y
	}

	return 0, 0
}

func (m *Map) checkPlayerRectAgainstTiles(playerRect image.Rectangle, check func(playerRect, tileRect image.Rectangle, isCollidable, isTouchable bool) bool) (bool, *tiles.Types) {
	for tile, positions := range m.positions {
		if !tile.Collidable && !tile.Touchable {
			continue
		}

		for _, position := range positions {
			x := int(position.X)
			y := int(position.Y)

			tileRect := image.Rect(x, y, x+ctypes.SpriteSize, y+ctypes.SpriteSize)

			ret := check(playerRect, tileRect, tile.Collidable, tile.Touchable)

			if ret {
				return ret, &tile
			}
		}
	}

	return false, nil
}

func (m *Map) IsColliding(x, y int) (bool, *tiles.Types) {
	playerRect := image.Rect(x, y, x+ctypes.SpriteSize, y+ctypes.SpriteSize)

	return m.checkPlayerRectAgainstTiles(playerRect, func(playerRect, tileRect image.Rectangle, isCollidable, _ bool) bool {
		if !isCollidable {
			return false
		}

		return playerRect.Overlaps(tileRect)
	})
}

func (m *Map) IsTouching(x, y int) (bool, *tiles.Types) {
	playerRect := image.Rect(x, y, x+ctypes.SpriteSize, y+ctypes.SpriteSize)

	return m.checkPlayerRectAgainstTiles(playerRect, func(playerRect, tileRect image.Rectangle, _, isTouchable bool) bool {
		if !isTouchable {
			return false
		}

		return playerRect.Overlaps(tileRect)
	})
}

func (m *Map) Draw(screen *ebiten.Image, tileset *tiles.Tiles) {
	for tileType, positions := range m.positions {
		for _, pos := range positions {
			switch tileType {
			case tiles.Typeses.GROUND_UL_TILE:
				tileset.Ground.DrawUL(screen, pos.X, pos.Y)
			case tiles.Typeses.GROUND_UM_TILE:
				tileset.Ground.DrawUM(screen, pos.X, pos.Y)
			case tiles.Typeses.GROUND_UR_TILE:
				tileset.Ground.DrawUR(screen, pos.X, pos.Y)

			case tiles.Typeses.GROUND_ML_TILE:
				tileset.Ground.DrawML(screen, pos.X, pos.Y)
			case tiles.Typeses.GROUND_MM_TILE:
				tileset.Ground.DrawMM(screen, pos.X, pos.Y)
			case tiles.Typeses.GROUND_MR_TILE:
				tileset.Ground.DrawMR(screen, pos.X, pos.Y)

			case tiles.Typeses.GROUND_BL_TILE:
				tileset.Ground.DrawBL(screen, pos.X, pos.Y)
			case tiles.Typeses.GROUND_BM_TILE:
				tileset.Ground.DrawBM(screen, pos.X, pos.Y)
			case tiles.Typeses.GROUND_BR_TILE:
				tileset.Ground.DrawBR(screen, pos.X, pos.Y)

			case tiles.Typeses.SPIKE_TILE:
				tileset.Spike.Draw(screen, pos.X, pos.Y)

			case tiles.Typeses.COIN_TILE:
				tileset.Coin.Draw(screen, pos.X, pos.Y)
			case tiles.Typeses.DIAMOND_TILE:
				tileset.Diamond.Draw(screen, pos.X, pos.Y)
			case tiles.Typeses.HEART_TILE:
				tileset.Heart.Draw(screen, pos.X, pos.Y)
			case tiles.Typeses.EMERALD_TILE:
				tileset.Emerald.Draw(screen, pos.X, pos.Y)

			case tiles.Typeses.DOOR_OPENED_TILE:
				tileset.Door.DrawOpened(screen, pos.X, pos.Y)
			case tiles.Typeses.DOOR_CLOSED_TILE:
				tileset.Door.DrawClosed(screen, pos.X, pos.Y)

			default:
			}
		}
	}
}

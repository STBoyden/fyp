package game

import (
	"fmt"
	"io"
	"os"
	"strings"

	"fyp/src/common/ctypes"
	"fyp/src/common/ctypes/tiles"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	maxMapWidth  = 40
	maxMapHeight = 30
)

type Map struct {
	content string
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
	m.content = string(content)

	return &m, nil
}

func (m *Map) GetSpawnPoint() (x, y float64) {
	for y, tileString := range strings.Split(m.content, "\n") {
		for x, c := range tileString {
			if c == 'S' {
				return float64(x) * ctypes.SpriteSizeF, float64(y) * ctypes.SpriteSizeF
			}
		}
	}

	return 0, 0
}

func (m *Map) Draw(screen *ebiten.Image, tileset *tiles.Tiles) {
	for line, tileString := range strings.Split(m.content, "\n") {
		for i, tileRune := range tileString {
			if line == maxMapHeight {
				break
			}

			if i != 0 && i%maxMapWidth == 0 {
				line++
				i = 0
			}

			x := float64(i) * ctypes.SpriteSizeF
			y := float64(line) * ctypes.SpriteSizeF

			switch tileRune {
			case '-':
				tileset.Ground.DrawUM(screen, x, y)
			case ']':
				tileset.Ground.DrawUR(screen, x, y)
			case '[':
				tileset.Ground.DrawUL(screen, x, y)

			case 'L':
				tileset.Ground.DrawML(screen, x, y)
			case 'M':
				tileset.Ground.DrawMM(screen, x, y)
			case 'R':
				tileset.Ground.DrawMR(screen, x, y)

			case '\\':
				tileset.Ground.DrawBL(screen, x, y)
			case '/':
				tileset.Ground.DrawBR(screen, x, y)
			case '_':
				tileset.Ground.DrawBM(screen, x, y)

			case 's':
				tileset.Spike.Draw(screen, x, y)
			case 'D':
				tileset.Door.DrawClosed(screen, x, y)
			case 'd':
				tileset.Door.DrawOpened(screen, x, y)

			default:
				tileset.Space.Draw(screen, x, y)
			}
		}
	}
}

//go:generate go run ../../internal/gen/generate-tiles/main.go -f ../../internal/gen/generate-tiles/spritesheet_data.yml

package ctypes

import (
	"errors"
	"fmt"
	"image"

	"fyp/resources"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	SpriteSize  = 16
	SpriteSizeF = float64(SpriteSize)
)

type Spritesheet struct {
	image    *ebiten.Image
	isLoaded bool
}

func (sheet *Spritesheet) Load() error {
	img, _, err := resources.GetImgTilemap()
	if err != nil {
		return err
	}

	sheet.image = img
	sheet.isLoaded = true

	return nil
}

func (sheet *Spritesheet) Sheet() (*ebiten.Image, error) {
	if !sheet.isLoaded {
		return nil, errors.New("spritesheet isn't loaded")
	}

	return sheet.image, nil
}

// gets player sprites from a given index 0-3.
func (sheet *Spritesheet) GetPlayer(index PlayerColour) ([]*ebiten.Image, error) {
	if !sheet.isLoaded {
		return nil, errors.New("spritesheet isn't loaded")
	}

	if index < PlayerMinColour {
		return nil, fmt.Errorf("index is too low, must be values between %d and %d (inclusive)", PlayerMinColour, PlayerMaxColour)
	} else if index > PlayerMaxColour {
		return nil, fmt.Errorf("index is too high, must be values between %d and %d (inclusive)", PlayerMinColour, PlayerMaxColour)
	}

	startY := 192 // the starting y position of the first character sprite
	offsetY := int(index) * SpriteSize
	endX := 112 // the ending x position of all the character sprites

	images := []*ebiten.Image{}

	for x := 0; x < endX; x += SpriteSize {
		rect := image.Rect(x, startY+offsetY, x+SpriteSize, startY+offsetY+SpriteSize)

		if subimage, ok := sheet.image.SubImage(rect).(*ebiten.Image); ok && subimage != nil {
			images = append(images, subimage)
		}
	}

	return images, nil
}

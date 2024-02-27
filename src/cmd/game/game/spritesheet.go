package game

import (
	"errors"
	"image"
	_ "image/png" // needed for the ebitenutil.NewImageFromFile

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Spritesheet struct {
	image    *ebiten.Image
	isLoaded bool
}

func (sheet *Spritesheet) Load() error {
	img, _, err := ebitenutil.NewImageFromFile("resources/images/tilemap_transparent_packed.png")
	if err != nil {
		return err
	}

	sheet.image = img
	sheet.isLoaded = true

	return nil
}

func (sheet *Spritesheet) GetPlayer(index int) ([]*ebiten.Image, error) {
	if !sheet.isLoaded {
		return nil, errors.New("Spritesheet isn't loaded")
	}

	if index < 0 {
		return nil, errors.New("index is too low, must be values between 0 and 3 (inclusive)")
	} else if index > 3 {
		return nil, errors.New("index is too high, must be values between 0 and 3 (inclusive)")
	}

	startY := 192 // the starting y position of the first character sprite
	offsetY := (spriteSize + 1) * index
	endX := 111 // the ending x position of all the character sprites

	images := []*ebiten.Image{}

	for x := 0; x < endX; x += spriteSize + 1 {
		rect := image.Rect(x, startY+offsetY, x+spriteSize, startY+offsetY+spriteSize)

		if subimage, ok := sheet.image.SubImage(rect).(*ebiten.Image); ok && subimage != nil {
			images = append(images, subimage)
		}
	}

	return images, nil
}

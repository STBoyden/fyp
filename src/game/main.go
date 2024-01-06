package main

import (
	logging "github.com/STBoyden/fyp/src/utils"
	ebiten "github.com/hajimehoshi/ebiten/v2"
)

type Game struct{}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {}

func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return outsideWidth, outsideHeight
}

func main() {
	game := &Game{}
	log := logging.NewClient()

	ebiten.SetWindowSize(1600, 900)
	ebiten.SetWindowTitle("Final Year Project")

	log.Info("Starting game...")
	if err := ebiten.RunGame(game); err != nil {
		log.Error(err.Error())
	}

	log.Info("Exited")
}

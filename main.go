// author: Jacky Boen

package main

import (
	"fmt"
	"os"

	"github.com/stuartdd/go_life"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

var winTitle string = "Go-SDL2 Render"
var winWidth, winHeight int32 = 900, 900
var btnBg = &sdl.Color{R: 0, G: 56, B: 0, A: 128}
var btnFg = &sdl.Color{R: 0, G: 255, B: 0, A: 255}
var btnHeight int32 = 70
var btnMarginTop int32 = 10

func run() int {
	var window *sdl.Window
	var renderer *sdl.Renderer
	var x int32
	var y int32
	var cellSize float32 = 5
	var cellScale float32 = 10
	var cellOffsetX float32 = float32(btnHeight)
	var cellOffsetY float32 = 0

	window, err := sdl.CreateWindow(winTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		winWidth, winHeight, sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		return 1
	}
	defer window.Destroy()

	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
		return 2
	}
	defer renderer.Destroy()

	// Load the font and ensure it is Closed() properly
	if err = ttf.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init the ttf font system: %s\n", err)
		return 1
	}
	defer ttf.Quit()
	font, err := ttf.OpenFont("Garuda-BoldOblique.ttf", 50)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load the font: %s\n", err)
		return 1
	}
	defer font.Close()

	running := true

	rle, err := go_life.NewRleFile("testdata/1234_synth.rle")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load RLE file: %s\n", err)
		return 1
	}

	lifeGen := go_life.NewLifeGen(func(lg *go_life.LifeGen) {}, go_life.RUN_FOR_EVER)
	lifeGen.AddCellsAtOffset(0, 0, go_life.COLOUR_MODE_MASK, rle.Coords())

	buttons := NewSDLWidgets(font)
	defer buttons.Destroy()
	arrows := NewSDLWidgets(nil)
	defer arrows.Destroy()

	b3 := NewSDLButton(330, btnMarginTop, 150, btnHeight, "Step", btnBg, btnFg, 10, func(b SDL_Widget, i1, i2 int32) bool {
		lifeGen.SetRunFor(1, nil)
		return true
	})
	b1 := NewSDLButton(10, btnMarginTop, 150, btnHeight, "Quit", btnBg, btnFg, 0, func(b SDL_Widget, i1, i2 int32) bool {
		running = false
		return true
	})
	b2 := NewSDLButton(170, btnMarginTop, 150, btnHeight, "Stop", btnBg, btnFg, 500, func(b SDL_Widget, i1, i2 int32) bool {
		bb := b.(*SDL_Button)
		if lifeGen.IsRunning() {
			lifeGen.SetRunFor(0, nil)
			bb.SetText("Start")
			b3.SetEnabled(true)
		} else {
			lifeGen.SetRunFor(go_life.RUN_FOR_EVER, nil)
			bb.SetText("Stop")
			b3.SetEnabled(false)
		}
		return true
	})
	b3.SetEnabled(false)
	buttons.Add(b1)
	buttons.Add(b2)
	buttons.Add(b3)

	a1 := NewSDLArrow(200, 300, 75, 50, 0, btnBg, btnFg, 0, nil)
	a2 := NewSDLArrow(200, 300, -75, 50, 0, btnBg, btnFg, 0, nil)
	a3 := NewSDLArrow(200, 300, 50, 75, 0, btnBg, btnFg, 0, nil)
	a4 := NewSDLArrow(200, 300, 50, -75, 0, btnBg, btnFg, 0, nil)
	arrows.Add(a1)
	arrows.Add(a2)
	arrows.Add(a3)
	arrows.Add(a4)

	for running {

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.MouseButtonEvent:
				if t.State == sdl.PRESSED {
					w := buttons.Inside(t.X, t.Y)
					if w != nil {
						go w.Click(t.X, t.Y)
					} else {
						cellOffsetX = float32(t.X)
						cellOffsetY = float32(t.Y)
					}
				}
			case *sdl.MouseWheelEvent:
				if t.X != 0 {
					fmt.Println("Mouse", t.Which, "wheel scrolled horizontally by", t.X)
				} else {
					fmt.Println("Mouse", t.Which, "wheel scrolled vertically by", t.Y)
				}
			}
		}

		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()
		renderer.SetDrawColor(0, 255, 255, 255)
		cell := lifeGen.GetRootCell()
		for cell != nil {
			x, y = cell.XY()
			renderer.DrawRectF(&sdl.FRect{X: cellOffsetX + float32(x)*cellScale, Y: cellOffsetY + float32(y)*cellScale, W: cellSize, H: cellSize})
			cell = cell.Next()
		}
		buttons.Draw(renderer)
		arrows.Draw(renderer)
		renderer.Present()
		lifeGen.NextGen()
	}

	return 0
}

func main() {
	os.Exit(run())
}

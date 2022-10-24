// author: Jacky Boen

package main

import (
	"fmt"
	"os"

	"github.com/stuartdd/go_life"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const (
	BUTTON_CLOSE int32 = iota
	BUTTON_STOP_START
	BUTTON_STEP
	ARROW_UP
	ARROW_DOWN
	ARROW_LEFT
	ARROW_RIGHT
)

var (
	winTitle            string = "Go-SDL2 Render"
	winWidth, winHeight int32  = 900, 900
	btnBg                      = &sdl.Color{R: 0, G: 56, B: 0, A: 128}
	btnFg                      = &sdl.Color{R: 0, G: 255, B: 0, A: 255}
	btnHeight           int32  = 70
	btnMarginTop        int32  = 10
	mouseX              int32  = 0
	mouseY              int32  = 0
	mouseOn                    = false
	cellSize            int32  = 5
	cellScale           int32  = 10
	cellOffsetX         int32  = btnHeight
	cellOffsetY         int32  = 0
	cellX               int32
	cellY               int32
	arrowPosX           int32 = 245
	arrowPosY           int32 = 180
	buttonWidth         int32 = 150
)

func run() int {
	var window *sdl.Window
	var renderer *sdl.Renderer

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
	font, err := ttf.OpenFont("resources/Garuda-BoldOblique.ttf", 50)
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

	buttons := NewSDLWidgetList(font)
	arrows := NewSDLWidgetList(nil)
	images := NewSDLWidgetList(nil)
	widgetGroup := NewWidgetGroup()
	widgetGroup.Add(buttons)
	widgetGroup.Add(arrows)
	widgetGroup.Add(images)
	widgetGroup.LoadTextures(renderer, "resources", map[string]string{
		"lem": "Apollo_LM.bmp",
	})
	defer widgetGroup.Destroy()

	b3 := NewSDLButton(330, btnMarginTop, buttonWidth, btnHeight, BUTTON_STEP, "Step", btnBg, btnFg, 10, func(b SDL_Widget, i1, i2 int32) bool {
		lifeGen.SetRunFor(1, nil)
		return true
	})
	b1 := NewSDLButton(10, btnMarginTop, buttonWidth, btnHeight, BUTTON_CLOSE, "Quit", btnBg, btnFg, 0, func(b SDL_Widget, i1, i2 int32) bool {
		running = false
		return true
	})
	b2 := NewSDLButton(170, btnMarginTop, buttonWidth, btnHeight, BUTTON_STOP_START, "Stop", btnBg, btnFg, 500, func(b SDL_Widget, i1, i2 int32) bool {
		bb := b.(*SDL_Button)
		if lifeGen.IsRunning() {
			lifeGen.SetRunFor(0, nil)
			bb.SetText("Start")
			b3.SetEnabled(true)
			arrows.SetVisible(true)
			mouseOn = true
		} else {
			lifeGen.SetRunFor(go_life.RUN_FOR_EVER, nil)
			bb.SetText("Stop")
			b3.SetEnabled(false)
			arrows.SetVisible(false)
			mouseOn = false
		}
		return true
	})
	b3.SetEnabled(false)
	buttons.Add(b1)
	buttons.Add(b2)
	buttons.Add(b3)

	arrowR := NewSDLArrow(arrowPosX, arrowPosY, 70, 50, ARROW_RIGHT, btnBg, btnFg, 0, func(s SDL_Widget, i1, i2 int32) bool {
		cellOffsetX = cellOffsetX + 100
		return true
	})
	arrowL := NewSDLArrow(arrowPosX, arrowPosY, -70, 50, ARROW_LEFT, btnBg, btnFg, 0, func(s SDL_Widget, i1, i2 int32) bool {
		cellOffsetX = cellOffsetX - 100
		return true
	})
	arrowD := NewSDLArrow(arrowPosX, arrowPosY, 50, 70, ARROW_DOWN, btnBg, btnFg, 0, func(s SDL_Widget, i1, i2 int32) bool {
		cellOffsetY = cellOffsetY + 100
		return true
	})
	arrowU := NewSDLArrow(arrowPosX, arrowPosY, 50, -70, ARROW_UP, btnBg, btnFg, 0, func(s SDL_Widget, i1, i2 int32) bool {
		cellOffsetY = cellOffsetY - 100
		return true
	})
	arrows.Add(arrowR)
	arrows.Add(arrowL)
	arrows.Add(arrowD)
	arrows.Add(arrowU)
	arrows.SetVisible(false)

	lem := NewSDLImage(400, 400, 0, 0, 999, "lem", btnBg, btnFg, 10, nil)
	images.Add(lem)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load 'lem': %s\n", err)
		return 1
	}

	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.MouseMotionEvent:
				mouseX = t.X
				mouseY = t.Y
			case *sdl.MouseButtonEvent:
				if t.State == sdl.PRESSED {
					w := widgetGroup.Inside(t.X, t.Y)
					if w != nil {
						go w.Click(t.X, t.Y)
					} else {
						cellOffsetX = t.X
						cellOffsetY = t.Y
					}
				}
			case *sdl.MouseWheelEvent:
				cellSize = cellSize + t.Y
				if cellSize < 1 {
					cellSize = 1
				}
				cellScale = cellSize * 2
			}
		}

		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()
		renderer.SetDrawColor(0, 255, 255, 255)
		cell := lifeGen.GetRootCell()
		for cell != nil {
			cellX, cellY = cell.XY()
			if cellSize > 5 {
				renderer.FillRect(&sdl.Rect{X: cellOffsetX + (cellX * cellScale), Y: cellOffsetY + (cellY * cellScale), W: cellSize, H: cellSize})
			} else {
				renderer.DrawRect(&sdl.Rect{X: cellOffsetX + (cellX * cellScale), Y: cellOffsetY + (cellY * cellScale), W: cellSize, H: cellSize})
			}
			cell = cell.Next()
		}
		lem.SetPosition(mouseX, mouseY)
		widgetGroup.Draw(renderer)
		if mouseOn {
			renderer.FillRect(&sdl.Rect{X: mouseX - (cellSize / 2), Y: mouseY - (cellSize / 2), W: cellSize, H: cellSize})
		}
		lifeGen.NextGen()
		renderer.Present()
	}
	return 0
}

func main() {
	os.Exit(run())
}

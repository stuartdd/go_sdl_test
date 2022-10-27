// author: Jacky Boen

package main

import (
	"fmt"
	"os"
	"path"

	go_life "github.com/stuartdd/go_life_engine"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const (
	BUTTON_CLOSE int32 = iota
	BUTTON_STOP_START
	BUTTON_STEP
	BUTTON_FASTER
	BUTTON_FASTEST
	BUTTON_SLOWER
	ARROW_UP
	ARROW_DOWN
	ARROW_LEFT
	ARROW_RIGHT
)

var (
	resources           string = "resources"
	winTitle            string = "Go-SDL2 Render"
	winWidth, winHeight int32  = 900, 900
	displayMode         sdl.DisplayMode
	btnBg                     = &sdl.Color{R: 0, G: 56, B: 0, A: 128}
	btnFg                     = &sdl.Color{R: 0, G: 255, B: 0, A: 255}
	btnHeight           int32 = 70
	btnMarginTop        int32 = 10
	btnWidth            int32 = 150
	btnGap              int32 = 10
	btnTopMarginHeight  int32 = btnHeight + (btnMarginTop * 2)
	mouseX              int32 = 0
	mouseY              int32 = 0
	mouseOn                   = false
	loopDelay           int   = 0
	lifeGen             *go_life.LifeGen
	cellSize            int32 = 5
	cellScale           int32 = 10
	cellOffsetX         int32 = btnHeight
	cellOffsetY         int32 = 0
	cellX               int32
	cellY               int32
	arrowPosX           int32 = 245
	arrowPosY           int32 = 180
)

func updateButtons(buttons, buttonsPaused, arrows *SDL_WidgetList) {
	l := buttons.ListWidgets()
	l = append(l, buttonsPaused.ListWidgets()...)
	for _, b := range l {
		switch b.GetId() {
		case BUTTON_FASTEST, BUTTON_FASTER:
			b.SetEnabled(loopDelay > 0)
		case BUTTON_STEP:
			b.SetVisible(lifeGen.GetRunFor() < 2)
		}
	}
	arrows.SetVisible(lifeGen.GetRunFor() < 2)
	x, y := buttons.ArrangeLR(btnGap, btnMarginTop, btnGap)
	buttonsPaused.ArrangeLR(x, y, btnGap)

}

func run() int {
	var window *sdl.Window
	var renderer *sdl.Renderer
	window, err := sdl.CreateWindow(winTitle, sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED,
		winWidth, winHeight, sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		return 1
	}
	defer window.Destroy()

	index, err := window.GetDisplayIndex()
	if err == nil {
		displayMode, err = sdl.GetCurrentDisplayMode(index)
		if err == nil {
			fmt.Printf("Window W:%d H:%d \n", displayMode.W, displayMode.H)
			window.SetSize(displayMode.W, displayMode.H)
		}
	}

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
	font, err := ttf.OpenFont(path.Join(resources, "buttonFont.ttf"), 50)
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

	lifeGen = go_life.NewLifeGen(func(lg *go_life.LifeGen) {}, go_life.RUN_FOR_EVER)
	lifeGen.AddCellsAtOffset(0, 0, go_life.COLOUR_MODE_MASK, rle.Coords())

	buttons := NewSDLWidgetList(font)
	buttonsPaused := NewSDLWidgetList(font)
	arrows := NewSDLWidgetList(nil)
	widgetGroup := NewWidgetGroup()
	widgetGroup.Add(buttons)
	widgetGroup.Add(buttonsPaused)
	widgetGroup.Add(arrows)

	// Load image resources
	err = widgetGroup.LoadTextures(renderer, resources, map[string]string{
		"lem":     "lem.png",
		"slower":  "slower.png",
		"faster":  "faster.png",
		"fastest": "fastest.png",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load textures: %s\n", err)
		return 1
	}
	defer widgetGroup.Destroy()

	btnStep := NewSDLButton(330, btnMarginTop, btnWidth, btnHeight, BUTTON_STEP, "Step", btnBg, btnFg, 10, func(b SDL_Widget, i1, i2 int32) bool {
		lifeGen.SetRunFor(1, nil)
		updateButtons(buttons, buttonsPaused, arrows)
		return true
	})

	btnSlower := NewSDLImage(400, 400, btnHeight, btnHeight, BUTTON_SLOWER, "slower", 0, btnBg, btnFg, 0, func(s SDL_Widget, i1, i2 int32) bool {
		loopDelay = loopDelay + 5
		updateButtons(buttons, buttonsPaused, arrows)
		return true
	})

	btnFaster := NewSDLImage(400, 400, btnHeight, btnHeight, BUTTON_FASTER, "faster", 0, btnBg, btnFg, 0, func(s SDL_Widget, i1, i2 int32) bool {
		loopDelay = loopDelay - 10
		if loopDelay < 0 {
			loopDelay = 0
		}
		updateButtons(buttons, buttonsPaused, arrows)
		return true
	})
	btnFastest := NewSDLImage(400, 400, btnHeight, btnHeight, BUTTON_FASTEST, "fastest", 0, btnBg, btnFg, 0, func(s SDL_Widget, i1, i2 int32) bool {
		loopDelay = 0
		updateButtons(buttons, buttonsPaused, arrows)
		return true
	})

	btnClose := NewSDLButton(10, btnMarginTop, btnWidth, btnHeight, BUTTON_CLOSE, "Quit", btnBg, btnFg, 0, func(b SDL_Widget, i1, i2 int32) bool {
		running = false
		return true
	})

	btnStop := NewSDLButton(170, btnMarginTop, btnWidth, btnHeight, BUTTON_STOP_START, "Stop", btnBg, btnFg, 500, func(b SDL_Widget, i1, i2 int32) bool {
		bb := b.(*SDL_Button)
		if lifeGen.IsRunning() {
			lifeGen.SetRunFor(0, nil)
			bb.SetText("Start")
			buttonsPaused.SetVisible(false)
			mouseOn = true
		} else {
			lifeGen.SetRunFor(go_life.RUN_FOR_EVER, nil)
			bb.SetText("Stop")
			buttonsPaused.SetVisible(true)
			mouseOn = false
		}
		updateButtons(buttons, buttonsPaused, arrows)
		return true
	})

	buttons.Add(btnClose)
	buttons.Add(btnStop)
	buttons.Add(NewSDLSeparator(0, 0, 10, btnHeight, 999, widgetColourDim(btnBg, false, 2)))
	buttons.Add(btnStep)
	buttonsPaused.Add(btnSlower)
	buttonsPaused.Add(btnFaster)
	buttonsPaused.Add(btnFastest)
	buttonsPaused.SetVisible(true)

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

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load 'lem': %s\n", err)
		return 1
	}
	updateButtons(buttons, buttonsPaused, arrows)
	for running {
		viewPort := renderer.GetViewport()
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
		renderer.SetDrawColor(0, 78, 0, 255)
		renderer.FillRect(&sdl.Rect{X: 0, Y: 0, W: viewPort.W, H: btnTopMarginHeight})
		renderer.SetDrawColor(0, 255, 255, 255)
		cell := lifeGen.GetRootCell()
		for cell != nil {
			cellX, cellY = cell.XY()
			y := cellOffsetY + (cellY * cellScale)
			if y > btnTopMarginHeight {
				if cellSize > 5 {
					renderer.FillRect(&sdl.Rect{X: cellOffsetX + (cellX * cellScale), Y: cellOffsetY + (cellY * cellScale), W: cellSize, H: cellSize})
				} else {
					renderer.DrawRect(&sdl.Rect{X: cellOffsetX + (cellX * cellScale), Y: cellOffsetY + (cellY * cellScale), W: cellSize, H: cellSize})
				}
			}
			cell = cell.Next()
		}
		widgetGroup.Draw(renderer)
		if mouseOn {
			renderer.FillRect(&sdl.Rect{X: mouseX - (cellSize / 2), Y: mouseY - (cellSize / 2), W: cellSize, H: cellSize})
		}
		lifeGen.NextGen()
		renderer.Present()
		sdl.Delay(uint32(loopDelay))
	}
	return 0
}

func main() {
	os.Exit(run())
}

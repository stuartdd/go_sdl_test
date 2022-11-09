// author: Stuart Davies

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
	BUTTON_NUM
	BUTTON_ZOOM_IN
	BUTTON_ZOOM_OUT
	BUTTON_LOAD_FILE
	LIST_TOP_LEFT
	LIST_PAUSED
	LIST_ARROWS
	LIST_TOP_RIGHT
	LABEL_GEN
	LABEL_SPEED
	LABEL_RELOAD
	PATH_ENTRY1
	PATH_ENTRY2
	ARROW_UP
	ARROW_DOWN
	ARROW_LEFT
	ARROW_RIGHT

	MAX_LOOP_DELAY  = 500
	MIN_LOOP_DELAY  = 0
	STEP_LOOP_DELAY = 50
)

var (
	resources           string = "resources"
	winTitle            string = "Go-SDL2 Render"
	winWidth, winHeight int32  = 900, 900
	displayMode         sdl.DisplayMode
	fontSize            int            = 50
	btnBg                              = &sdl.Color{R: 0, G: 56, B: 0, A: 128}
	btnFg                              = &sdl.Color{R: 0, G: 255, B: 0, A: 255}
	mouseData           *SDL_MouseData = &SDL_MouseData{}
	btnHeight           int32          = 70
	btnMarginTop        int32          = 10
	btnWidth            int32          = 150
	btnGap              int32          = 10
	btnTopMarginHeight  int32          = btnHeight + (btnMarginTop * 2)
	mouseOn                            = false
	loopDelay           uint32         = 0
	cellSize            int32          = 3
	gridSize            int32          = 5
	cellOffsetX         int32          = 0
	cellOffsetY         int32          = 0
	cellX               int32
	cellY               int32
	arrowPosX           int32 = 245
	arrowPosY           int32 = 245
	lifeGen             *go_life.LifeGen
	rleFile             *go_life.RLE
)

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
	font, err := ttf.OpenFont(path.Join(resources, "buttonFont.ttf"), fontSize)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load the font: %s\n", err)
		return 1
	}
	defer font.Close()

	running := true

	err = loadRleFile("testdata/1234_synth.rle")
	if err != nil {
		fmt.Printf("Unable to load file %e", err)
	}
	viewPort := renderer.GetViewport()
	cellOffsetX, cellOffsetY = centerOnXY(viewPort.W/2, viewPort.H/2, lifeGen)

	widgetGroup := NewWidgetGroup(font)
	defer widgetGroup.Destroy()

	arrows := widgetGroup.NewWidgetSubGroup(nil, LIST_ARROWS)
	buttonsTL := widgetGroup.NewWidgetSubGroup(nil, LIST_TOP_LEFT)
	buttonsPaused := widgetGroup.NewWidgetSubGroup(nil, LIST_PAUSED)
	buttonsTR := widgetGroup.NewWidgetSubGroup(nil, LIST_TOP_RIGHT)

	// Load image resources
	err = widgetGroup.LoadTexturesFromFileMap(renderer, resources, map[string]string{
		"lem":      "lem.png",
		"slower":   "slower.png",
		"faster":   "faster.png",
		"fastest":  "fastest.png",
		"zoomin":   "zoom-in.png",
		"zoomout":  "zoom-out.png",
		"fileload": "file-load.png",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load file textures: %s\n", err)
		return 1
	}

	err = widgetGroup.LoadTexturesFromStringMap(renderer, map[string]string{
		"numbers": "0123456789",
	}, font, btnFg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load text textures: %s\n", err)
		return 1
	}

	btnClose := NewSDLButton(0, 0, btnWidth, btnHeight, BUTTON_CLOSE, "Quit", btnBg, btnFg, 0, func(b SDL_Widget, i1, i2 int32) bool {
		running = false
		return true
	})

	btnSlower := NewSDLImage(0, 0, btnHeight, btnHeight, BUTTON_SLOWER, "slower", 0, 1, btnBg, btnFg, 0, func(s SDL_Widget, i1, i2 int32) bool {
		loopDelay = loopDelay + STEP_LOOP_DELAY
		if loopDelay > MAX_LOOP_DELAY {
			loopDelay = MAX_LOOP_DELAY
		}
		return true
	})

	btnFaster := NewSDLImage(0, 0, btnHeight, btnHeight, BUTTON_FASTER, "faster", 0, 1, btnBg, btnFg, 0, func(s SDL_Widget, i1, i2 int32) bool {
		if loopDelay >= STEP_LOOP_DELAY {
			loopDelay = loopDelay - STEP_LOOP_DELAY
		}
		return true
	})

	btnFastest := NewSDLImage(0, 0, btnHeight, btnHeight, BUTTON_FASTEST, "fastest", 0, 1, btnBg, btnFg, 0, func(s SDL_Widget, i1, i2 int32) bool {
		loopDelay = MIN_LOOP_DELAY
		return true
	})

	btnZoomIn := NewSDLImage(0, 0, btnHeight, btnHeight, BUTTON_ZOOM_IN, "zoomin", 0, 1, nil, btnFg, 0, func(s SDL_Widget, i1, i2 int32) bool {
		widgetGroup.Scale(1.1)
		_, h := btnClose.GetSize()
		btnTopMarginHeight = h + +(btnMarginTop * 2)
		updateButtons(renderer, widgetGroup)
		return true
	})

	btnZoomOut := NewSDLImage(0, 0, btnHeight, btnHeight, BUTTON_ZOOM_IN, "zoomout", 0, 1, nil, btnFg, 0, func(s SDL_Widget, i1, i2 int32) bool {
		widgetGroup.Scale(0.9)
		_, h := btnClose.GetSize()
		btnTopMarginHeight = h + +(btnMarginTop * 2)
		updateButtons(renderer, widgetGroup)
		return true
	})

	labelGen := NewSDLLabel(0, 0, 290, btnHeight, LABEL_GEN, "Gen:0", ALIGN_LEFT, btnBg, btnFg)
	labelSpeed := NewSDLLabel(0, 0, 270, btnHeight, LABEL_SPEED, "Delay:0ms", ALIGN_LEFT, btnBg, btnFg)

	var loadFile *SDL_Image
	pathEntry1 := NewSDLEntry(0, 0, 500, btnHeight, PATH_ENTRY1, rleFile.Filename(), btnBg, btnFg, func(old, new string, t TEXT_CHANGE_TYPE) (string, error) {
		fmt.Printf("OnChange old:'%s' new:'%s', type:%d\n", old, new, t)
		_, err := os.Stat(new)
		loadFile.SetEnabled(err == nil)
		updateButtons(renderer, widgetGroup)
		return new, err
	})

	loadFile = NewSDLImage(0, 0, btnHeight, btnHeight, BUTTON_LOAD_FILE, "fileload", 0, 1, btnBg, btnFg, 0, func(s SDL_Widget, i1, i2 int32) bool {
		loadRleFile(pathEntry1.GetText())
		return true
	})

	btnStep := NewSDLButton(0, 0, btnWidth, btnHeight, BUTTON_STEP, "Step", btnBg, btnFg, 10, func(b SDL_Widget, i1, i2 int32) bool {
		lifeGen.SetRunFor(1, nil)
		return true
	})

	btnStop := NewSDLButton(0, 0, btnWidth+30, btnHeight, BUTTON_STOP_START, "Stop", btnBg, btnFg, 500, func(b SDL_Widget, i1, i2 int32) bool {
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
		return true
	})

	arrowR := NewSDLShapeArrowRight(arrowPosX, arrowPosY, 100, 100, ARROW_RIGHT, btnBg, btnFg, func(s SDL_Widget, i1, i2 int32) bool {
		cellOffsetX = cellOffsetX + 100
		return true
	})
	arrowD := NewSDLShapeArrowRight(arrowPosX, arrowPosY, 100, 100, ARROW_DOWN, btnBg, btnFg, func(s SDL_Widget, i1, i2 int32) bool {
		cellOffsetY = cellOffsetY + 100
		return true
	})
	arrowD.Rotate(90)
	arrowL := NewSDLShapeArrowRight(arrowPosX, arrowPosY, 100, 100, ARROW_LEFT, btnBg, btnFg, func(s SDL_Widget, i1, i2 int32) bool {
		cellOffsetX = cellOffsetX - 100
		return true
	})
	arrowL.Rotate(180)
	arrowU := NewSDLShapeArrowRight(arrowPosX, arrowPosY, 100, 100, ARROW_UP, btnBg, btnFg, func(s SDL_Widget, i1, i2 int32) bool {
		cellOffsetY = cellOffsetY - 100
		return true
	})
	arrowU.Rotate(-90)

	buttonsTL.Add(btnClose)
	buttonsTL.Add(btnStop)
	buttonsTL.Add(labelGen)
	buttonsTL.Add(NewSDLSeparator(0, 0, 10, btnHeight, 999, widgetColourDim(btnBg, false, 2)))
	buttonsTL.Add(btnStep)
	buttonsTR.Add(btnZoomIn)
	buttonsTR.Add(btnZoomOut)

	buttonsPaused.Add(btnSlower)
	buttonsPaused.Add(btnFaster)
	buttonsPaused.Add(btnFastest)
	buttonsPaused.Add(labelSpeed)
	buttonsPaused.Add(NewSDLSeparator(0, 0, 10, btnHeight, 999, widgetColourDim(btnBg, false, 2)))
	buttonsPaused.Add(NewSDLLabel(0, 0, 250, btnHeight, LABEL_RELOAD, "Reload", ALIGN_LEFT, btnBg, nil))
	buttonsPaused.Add(loadFile)
	buttonsPaused.Add(pathEntry1)

	arrows.Add(arrowR)
	arrows.Add(arrowD)
	arrows.Add(arrowL)
	arrows.Add(arrowU)

	buttonsPaused.SetVisible(true)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load 'lem': %s\n", err)
		return 1
	}

	go func() {
		for running {
			labelGen.SetText(fmt.Sprintf("Gen:%d", lifeGen.GetGenerationCount()))
			updateButtons(renderer, widgetGroup)
			sdl.Delay(300)
		}
	}()

	go func() {
		for running {
			sdl.Delay(loopDelay)
			lifeGen.NextGen()
		}
	}()

	for running {
		viewPort := renderer.GetViewport()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.TextInputEvent:
				for _, c := range t.GetText() {
					widgetGroup.KeyPress(int(c), false, true)
				}
			case *sdl.KeyboardEvent:
				ks := t.Keysym.Sym
				if t.State == sdl.PRESSED {
					if ks == sdl.K_ESCAPE {
						running = false
					} else {
						widgetGroup.KeyPress(int(ks), true, true)
					}
				} else {
					widgetGroup.KeyPress(int(ks), true, false)
				}
			case *sdl.MouseMotionEvent:
				w := widgetGroup.Inside(t.X, t.Y)
				if w != nil && w.GetWidgetId() == mouseData.GetWidgetId() {
					if t.State == sdl.PRESSED {
						mouseData.SetDragging(true)
						mouseData.SetXY(t.X, t.Y)
						w.Click(mouseData)
					}
				} else {
					mouseData.SetDragging(false)
					mouseData.SetXY(t.X, t.Y)
					widgetGroup.ClearSelection()
				}
			case *sdl.MouseButtonEvent:
				x := t.X
				y := t.Y
				w := widgetGroup.Inside(x, y)
				if w != nil {
					if t.State == sdl.PRESSED {
						widgetId := w.GetWidgetId()
						mouseData.SetXY(x, y)
						mouseData.SetButtons(t.Button)
						mouseData.SetWidgetId(widgetId)
						widgetGroup.SetFocus(widgetId)
						w.Click(mouseData)
					} else {
						if mouseData.IsDragging() {
							mouseData.SetDragged(true)
							w.Click(mouseData)
							mouseData.SetDragged(false)
						}
					}
				} else {
					if widgetGroup.GetFocused() == nil {
						cellOffsetX, cellOffsetY = centerOnXY(x, y, lifeGen)
					} else {
						widgetGroup.ClearFocus()
					}
				}
			case *sdl.MouseWheelEvent:
				zoomGrid(t.Y)
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
			x := (cellOffsetX + cellX) * gridSize
			y := (cellOffsetY + cellY) * gridSize
			if cellSize < 2 {
				renderer.DrawPoint(x, y)
			} else {
				renderer.FillRect(&sdl.Rect{X: x, Y: y, W: cellSize, H: cellSize})
			}
			cell = cell.Next()
		}
		widgetGroup.Draw(renderer)
		if mouseOn {
			renderer.SetDrawColor(0, 0, 255, 255)
			renderer.DrawRect(&sdl.Rect{X: mouseData.GetX() - (30 / 2), Y: mouseData.GetY() - (30 / 2), W: 30, H: 30})
		}
		renderer.Present()
		sdl.Delay(20)
	}
	return 0
}

func main() {
	os.Exit(run())
}

func centerOnXY(x, y int32, lg *go_life.LifeGen) (int32, int32) {
	ax, ay := average(lg)
	return (x / gridSize) - ax, (y / gridSize) - ay
}

func average(lg *go_life.LifeGen) (int32, int32) {
	var at int32 = 40
	var ax1, ay1, ax2, ay2, x, y, tot, sumX, sumY int32 = 0, 0, 0, 0, 0, 0, 0, 0, 0
	cell := lg.GetRootCell()
	for cell != nil {
		tot = tot + 1
		x, y = cell.XY()
		ax1 = sumX / tot
		sumX = sumX + x
		ax2 = sumX / tot
		if ax2 < (ax1-at) || ax2 > (ax1+at) {
			sumX = sumX - x
		}
		ay1 = sumY / tot
		sumY = sumY + y
		ay2 = sumX / tot
		if ay2 < (ay1-at) || ay2 > (ay1+at) {
			sumY = sumY - y
		}
		cell = cell.Next()
	}
	return sumX / tot, sumY / tot
}

func updateButtons(renderer *sdl.Renderer, wg *SDL_WidgetGroup) {
	wl := wg.AllWidgets()
	for _, w := range wl {
		ww := *w
		switch (*w).GetWidgetId() {
		case BUTTON_FASTEST, BUTTON_FASTER:
			ww.SetEnabled(loopDelay > MIN_LOOP_DELAY)
		case BUTTON_SLOWER:
			ww.SetEnabled(loopDelay < MAX_LOOP_DELAY)
		case BUTTON_STEP:
			ww.SetVisible(lifeGen.GetRunFor() < 2)
		case LABEL_SPEED:
			ww.(*SDL_Label).SetText(fmt.Sprintf("Speed:%d", (MAX_LOOP_DELAY/50)-(loopDelay/50)))
		case PATH_ENTRY1:
			x, _ := ww.GetPosition()
			ww.SetSize(renderer.GetViewport().W-x-300, -1)

		}
	}

	var x, y int32 = 0, 0
	ll := wg.AllSubGroups()
	for _, l := range ll {
		switch l.GetId() {
		case LIST_ARROWS:
			l.SetEnable(lifeGen.GetRunFor() < 2)
		case LIST_TOP_LEFT:
			x, y = l.ArrangeLR(btnGap, btnMarginTop, btnGap)
			wg.GetWidgetSubGroup(LIST_PAUSED).ArrangeLR(x, y, btnGap)
		case LIST_TOP_RIGHT:
			l.ArrangeRL(renderer.GetViewport().W-btnGap, btnMarginTop, btnGap)
		}
	}
}

func zoomGrid(y int32) {
	gridSize = gridSize + y
	cellSize = gridSize - 2
	if gridSize < 3 {
		gridSize = 3
	}
	if cellSize < 2 {
		cellSize = 1
	}
	fmt.Printf("Size = %d Scale = %d\n", cellSize, gridSize)
}

func loadRleFile(filename string) error {
	var err error
	rleFile, err = go_life.NewRleFile(filename)
	if err != nil {
		return err
	}
	lifeGen = go_life.NewLifeGen(func(lg *go_life.LifeGen) {}, go_life.RUN_FOR_EVER)
	lifeGen.AddCellsAtOffset(0, 0, go_life.COLOUR_MODE_MASK, rleFile.Coords())
	return nil
}

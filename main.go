// author: Stuart Davies

package main

import (
	"fmt"
	"os"
	"path"

	go_life "github.com/stuartdd/go_life_engine"
	widgets "github.com/stuartdd/go_sdl_widget"
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
	STATUS_BOTTOM_LEFT
	LABEL_GEN
	LABEL_SPEED
	LABEL_LOG
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
	fontSize            int                    = 50
	btnFg                                      = &sdl.Color{R: 0, G: 255, B: 0, A: 255}
	mouseData           *widgets.SDL_MouseData = &widgets.SDL_MouseData{}
	btnHeight           int32                  = 70
	btnMarginTop        int32                  = 10
	btnWidth            int32                  = 200
	btnGap              int32                  = 10
	btnTopMarginHeight  int32                  = btnHeight + (btnMarginTop * 2)
	mouseOn                                    = false
	loopDelay           uint32                 = 0
	cellSize            int32                  = 3
	gridSize            int32                  = 5
	cellOffsetX         int32                  = 0
	cellOffsetY         int32                  = 0
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

	widgets.GetResourceInstance().SetFont(font)
	running := true

	err = loadRleFile("testdata/1234_synth.rle")
	if err != nil {
		fmt.Printf("Unable to load file %e", err)
	}
	viewPort := renderer.GetViewport()
	cellOffsetX, cellOffsetY = centerOnXY(viewPort.W/2, viewPort.H/2, lifeGen)

	widgetGroup := widgets.NewWidgetGroup(font)
	defer widgetGroup.Destroy()

	arrows := widgetGroup.NewWidgetSubGroup(nil, LIST_ARROWS)
	buttonsTL := widgetGroup.NewWidgetSubGroup(nil, LIST_TOP_LEFT)
	buttonsPaused := widgetGroup.NewWidgetSubGroup(nil, LIST_PAUSED)
	buttonsTR := widgetGroup.NewWidgetSubGroup(nil, LIST_TOP_RIGHT)
	statusBL := widgetGroup.NewWidgetSubGroup(nil, STATUS_BOTTOM_LEFT)

	// Load image resources
	err = widgets.GetResourceInstance().AddTexturesFromFileMap(renderer, resources, map[string]string{
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

	err = widgets.GetResourceInstance().AddTexturesFromStringMap(renderer, map[string]string{
		"numbers": "0123456789",
	}, font, btnFg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load text textures: %s\n", err)
		return 1
	}

	btnClose := widgets.NewSDLButton(0, 0, btnWidth, btnHeight, BUTTON_CLOSE, "Quit", widgets.WIDGET_STYLE_BORDER_AND_BG, 0, func(b widgets.SDL_Widget, i1, i2 int32) bool {
		running = false
		return true
	})
	btnSlower := widgets.NewSDLImage(0, 0, btnHeight, btnHeight, BUTTON_SLOWER, "slower", 0, 1, widgets.WIDGET_STYLE_BORDER_AND_BG, 0, func(s widgets.SDL_Widget, i1, i2 int32) bool {
		loopDelay = loopDelay + STEP_LOOP_DELAY
		if loopDelay > MAX_LOOP_DELAY {
			loopDelay = MAX_LOOP_DELAY
		}
		return true
	})

	btnFaster := widgets.NewSDLImage(0, 0, btnHeight, btnHeight, BUTTON_FASTER, "faster", 0, 1, widgets.WIDGET_STYLE_BORDER_AND_BG, 0, func(s widgets.SDL_Widget, i1, i2 int32) bool {
		if loopDelay >= STEP_LOOP_DELAY {
			loopDelay = loopDelay - STEP_LOOP_DELAY
		}
		return true
	})

	btnFastest := widgets.NewSDLImage(0, 0, btnHeight, btnHeight, BUTTON_FASTEST, "fastest", 0, 1, widgets.WIDGET_STYLE_BORDER_AND_BG, 0, func(s widgets.SDL_Widget, i1, i2 int32) bool {
		loopDelay = MIN_LOOP_DELAY
		return true
	})

	btnZoomIn := widgets.NewSDLImage(0, 0, btnHeight, btnHeight, BUTTON_ZOOM_IN, "zoomin", 0, 1, widgets.WIDGET_STYLE_BORDER_AND_BG, 0, func(s widgets.SDL_Widget, i1, i2 int32) bool {
		widgetGroup.Scale(1.1)
		_, h := btnClose.GetSize()
		btnTopMarginHeight = h + +(btnMarginTop * 2)
		updateButtons(renderer, widgetGroup)
		return true
	})

	btnZoomOut := widgets.NewSDLImage(0, 0, btnHeight, btnHeight, BUTTON_ZOOM_IN, "zoomout", 0, 1, widgets.WIDGET_STYLE_BORDER_AND_BG, 0, func(s widgets.SDL_Widget, i1, i2 int32) bool {
		widgetGroup.Scale(0.9)
		_, h := btnClose.GetSize()
		btnTopMarginHeight = h + +(btnMarginTop * 2)
		updateButtons(renderer, widgetGroup)
		return true
	})

	labelGen := widgets.NewSDLLabel(0, 0, 350, btnHeight, LABEL_GEN, "Gen:0", widgets.ALIGN_LEFT, widgets.WIDGET_STYLE_NONE)
	labelSpeed := widgets.NewSDLLabel(0, 0, 350, btnHeight, LABEL_SPEED, "Delay:0ms", widgets.ALIGN_LEFT, widgets.WIDGET_STYLE_NONE)
	labelLog := widgets.NewSDLLabel(0, viewPort.H-btnHeight, viewPort.W, btnHeight, LABEL_LOG, "Delay:0ms", widgets.ALIGN_LEFT, widgets.WIDGET_STYLE_DRAW_BORDER)
	labelLog.SetForeground(&sdl.Color{R: 100, G: 100, B: 100, A: 255})
	labelLog.SetBorderColour(&sdl.Color{R: 100, G: 100, B: 100, A: 255})
	var loadFile *widgets.SDL_Button

	pathEntry1 := widgets.NewSDLEntry(0, 0, 500, btnHeight, PATH_ENTRY1, rleFile.Filename(), widgets.WIDGET_STYLE_BORDER_AND_BG, func(old, new string, t widgets.TEXT_CHANGE_TYPE) (string, error) {
		if t == widgets.TEXT_CHANGE_SELECTED {
			return new, err
		}
		_, err := os.Stat(new)
		loadFile.SetEnabled(err == nil)
		updateButtons(renderer, widgetGroup)
		return new, err
	})

	loadFile = widgets.NewSDLButton(0, 0, btnWidth, btnHeight, BUTTON_LOAD_FILE, "Load", widgets.WIDGET_STYLE_BORDER_AND_BG, 500, func(s widgets.SDL_Widget, i1, i2 int32) bool {
		loadRleFile(pathEntry1.GetText())
		return true
	})

	btnStep := widgets.NewSDLButton(0, 0, btnWidth, btnHeight, BUTTON_STEP, "Step", widgets.WIDGET_STYLE_BORDER_AND_BG, 10, func(b widgets.SDL_Widget, i1, i2 int32) bool {
		lifeGen.SetRunFor(1, nil)
		return true
	})

	btnStop := widgets.NewSDLButton(0, 0, btnWidth+30, btnHeight, BUTTON_STOP_START, "Stop", widgets.WIDGET_STYLE_BORDER_AND_BG, 500, func(b widgets.SDL_Widget, i1, i2 int32) bool {
		bb := b.(*widgets.SDL_Button)
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

	arrowR := widgets.NewSDLShapeArrowRight(arrowPosX, arrowPosY, 100, 100, ARROW_RIGHT, widgets.WIDGET_STYLE_BORDER_AND_BG, func(s widgets.SDL_Widget, i1, i2 int32) bool {
		cellOffsetX = cellOffsetX + 100
		return true
	})
	arrowD := widgets.NewSDLShapeArrowRight(arrowPosX, arrowPosY, 100, 100, ARROW_DOWN, widgets.WIDGET_STYLE_BORDER_AND_BG, func(s widgets.SDL_Widget, i1, i2 int32) bool {
		cellOffsetY = cellOffsetY + 100
		return true
	})
	arrowD.Rotate(90)
	arrowL := widgets.NewSDLShapeArrowRight(arrowPosX, arrowPosY, 100, 100, ARROW_LEFT, widgets.WIDGET_STYLE_BORDER_AND_BG, func(s widgets.SDL_Widget, i1, i2 int32) bool {
		cellOffsetX = cellOffsetX - 100
		return true
	})
	arrowL.Rotate(180)
	arrowU := widgets.NewSDLShapeArrowRight(arrowPosX, arrowPosY, 100, 100, ARROW_UP, widgets.WIDGET_STYLE_BORDER_AND_BG, func(s widgets.SDL_Widget, i1, i2 int32) bool {
		cellOffsetY = cellOffsetY - 100
		return true
	})
	arrowU.Rotate(-90)

	buttonsTL.Add(btnClose)
	buttonsTL.Add(btnStop)
	buttonsTL.Add(labelGen)
	buttonsTL.Add(widgets.NewSDLSeparator(0, 0, 20, btnHeight, 999, widgets.WIDGET_STYLE_DRAW_BORDER))
	buttonsTL.Add(btnStep)
	buttonsTR.Add(btnZoomIn)
	buttonsTR.Add(btnZoomOut)

	buttonsPaused.Add(btnSlower)
	buttonsPaused.Add(btnFaster)
	buttonsPaused.Add(btnFastest)
	buttonsPaused.Add(labelSpeed)
	buttonsPaused.Add(widgets.NewSDLSeparator(0, 0, 20, btnHeight, 999, widgets.WIDGET_STYLE_DRAW_BORDER))
	buttonsPaused.Add(loadFile)
	buttonsPaused.Add(pathEntry1)

	arrows.Add(arrowR)
	arrows.Add(arrowD)
	arrows.Add(arrowL)
	arrows.Add(arrowU)

	statusBL.Add(labelLog)

	buttonsPaused.SetVisible(true)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load 'lem': %s\n", err)
		return 1
	}

	go func() {
		for running {
			labelGen.SetText(fmt.Sprintf("Gen:%d", lifeGen.GetGenerationCount()))
			updateButtons(renderer, widgetGroup)
			sdl.Delay(500)
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
						mouseData.SetClickCount(1)
						w.Click(mouseData)
					}
				} else {
					mouseData.SetDragging(false)
					mouseData.SetXY(t.X, t.Y)
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
						mouseData.SetClickCount(int(t.Clicks))
						widgetGroup.SetFocused(widgetId)
						go w.Click(mouseData)
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

func updateButtons(renderer *sdl.Renderer, wg *widgets.SDL_WidgetGroup) {
	viewport := renderer.GetViewport()

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
			ww.(*widgets.SDL_Label).SetText(fmt.Sprintf("Speed:%d", (MAX_LOOP_DELAY/50)-(loopDelay/50)))
		case PATH_ENTRY1:
			x, _ := ww.GetPosition()
			ww.SetSize(viewport.W-x-300, -1)
		case LABEL_LOG:
			_, h := ww.GetSize()
			ww.SetPosition(0, viewport.H-h)
			ww.SetSize(renderer.GetViewport().W, -1)
		}
	}

	var x, y int32 = 0, 0
	ll := wg.AllSubGroups()
	for _, l := range ll {
		switch l.GetId() {
		case LIST_ARROWS:
			l.SetVisible(lifeGen.GetRunFor() < 2)
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

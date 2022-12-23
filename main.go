// author: Stuart Davies

package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	go_life "github.com/stuartdd/go_life_engine"
	widgets "github.com/stuartdd/go_sdl_widget"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const (
	BUTTON_CLOSE int32 = iota + 10000
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
	LIST_RUNNING
	LIST_PAUSED
	LIST_ARROWS
	LIST_TOP_RIGHT
	LIST_FILES
	STATUS_BOTTOM_LEFT
	LABEL_GEN
	LABEL_SPEED
	LABEL_LOG
	PATH_ENTRY
	ARROW_UP
	ARROW_DOWN
	ARROW_LEFT
	ARROW_RIGHT

	MAX_LOOP_DELAY  = 500
	MIN_LOOP_DELAY  = 0
	STEP_LOOP_DELAY = 50

	DIST_BTN_HEIGHT = 70
	DIST_BTN_WIDTH  = 200
	DIST_BTN_SPACER = 10
	DIST_OFFSET_TOP = 10
)

var (
	resources           string = "resources"
	winTitle            string = "Go-SDL2 Render"
	winWidth, winHeight int32  = 900, 900
	displayMode         sdl.DisplayMode
	statusLabel         *widgets.SDL_Label
	statusGroup         *widgets.SDL_WidgetSubGroup
	fontSize            int                    = 50
	mouseData           *widgets.SDL_MouseData = &widgets.SDL_MouseData{}

	mouseOn            = false
	cellSize     int32 = 3
	gridSize     int32 = 5
	cellOffsetX  int32 = 0
	cellOffsetY  int32 = 0
	cellX        int32
	cellY        int32
	arrowPosX    int32 = 245
	arrowPosY    int32 = 245
	lifeGen      *go_life.LifeGen
	lifeGenSpeed uint64
	lifeGenTime  uint64
	loopDelay    uint64 = 0
	rleFile      *go_life.RLE

	// Set and scaled via the scaleEverything method
	viewport           sdl.Rect
	btnHeight          int32 = DIST_BTN_HEIGHT
	btnMarginTopS      int32 = DIST_OFFSET_TOP
	btnTopMarginHeight int32 = DIST_BTN_HEIGHT + DIST_OFFSET_TOP + DIST_OFFSET_TOP
	btnWidth           int32 = DIST_BTN_WIDTH
	btnGap             int32 = DIST_BTN_SPACER
)

func scaleEverything(r *sdl.Renderer, s float32, wg *widgets.SDL_WidgetGroup) {
	btnHeight = int32(float32(btnHeight) * s)
	btnMarginTopS = int32(float32(btnMarginTopS) * s)
	btnTopMarginHeight = int32(float32(btnTopMarginHeight) * s)
	btnWidth = int32(float32(btnWidth) * s)
	btnGap = int32(float32(btnGap) * s)
	if r != nil {
		viewport = r.GetViewport()
	}
	if wg != nil {
		wg.Scale(s)
	}
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
		} else {
			fmt.Fprintf(os.Stderr, "Failed to obtain displa mode: %s\n", err)
			return 2
		}
	}

	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
		return 3
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
	viewport = renderer.GetViewport()
	cellOffsetX, cellOffsetY = centerOnXY(viewport.W/2, viewport.H/2, lifeGen)

	widgetGroup := widgets.NewWidgetGroup(font)
	defer widgetGroup.Destroy()

	scaleEverything(renderer, 1.0, widgetGroup)

	buttonsRunning := widgetGroup.NewWidgetSubGroup(0, 100, 0, 0, LIST_RUNNING, widgets.WIDGET_STYLE_NONE)
	buttonsPaused := widgetGroup.NewWidgetSubGroup(0, 0, 0, 0, LIST_PAUSED, widgets.WIDGET_STYLE_NONE)
	statusGroup = widgetGroup.NewWidgetSubGroup(0, 0, 0, 0, STATUS_BOTTOM_LEFT, widgets.WIDGET_STYLE_NONE)
	buttonsTR := widgetGroup.NewWidgetSubGroup(0, 0, 0, 0, LIST_TOP_RIGHT, widgets.WIDGET_STYLE_NONE)
	buttonsTL := widgetGroup.NewWidgetSubGroup(0, 0, 0, 0, LIST_TOP_LEFT, widgets.WIDGET_STYLE_NONE)
	arrows := widgetGroup.NewWidgetSubGroup(0, 0, 0, 0, LIST_ARROWS, widgets.WIDGET_STYLE_NONE)

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

	btnClose := widgets.NewSDLButton(0, 0, btnWidth, btnHeight, BUTTON_CLOSE, "Quit", widgets.WIDGET_STYLE_DRAW_BORDER_AND_BG, 0, func(s string, b, i1, i2 int32) bool {
		running = false
		return true
	})

	btnSlower := widgets.NewSDLImage(0, 0, btnHeight, btnHeight, BUTTON_SLOWER, "slower", 0, 1, widgets.WIDGET_STYLE_DRAW_BORDER_AND_BG, 0, func(s string, b, i1, i2 int32) bool {
		loopDelay = loopDelay + STEP_LOOP_DELAY
		if loopDelay > MAX_LOOP_DELAY {
			loopDelay = MAX_LOOP_DELAY
		}
		return true
	})

	btnFaster := widgets.NewSDLImage(0, 0, btnHeight, btnHeight, BUTTON_FASTER, "faster", 0, 1, widgets.WIDGET_STYLE_DRAW_BORDER_AND_BG, 0, func(s string, b, i1, i2 int32) bool {
		if loopDelay >= STEP_LOOP_DELAY {
			loopDelay = loopDelay - STEP_LOOP_DELAY
		}
		return true
	})

	btnFastest := widgets.NewSDLImage(0, 0, btnHeight, btnHeight, BUTTON_FASTEST, "fastest", 0, 1, widgets.WIDGET_STYLE_DRAW_BORDER_AND_BG, 0, func(s string, b, i1, i2 int32) bool {
		loopDelay = MIN_LOOP_DELAY
		return true
	})

	btnZoomIn := widgets.NewSDLImage(0, 0, btnHeight, btnHeight, BUTTON_ZOOM_IN, "zoomin", 0, 1, widgets.WIDGET_STYLE_DRAW_BORDER_AND_BG, 0, func(s string, b, i1, i2 int32) bool {
		scaleEverything(renderer, 1.1, widgetGroup)
		return true
	})

	btnZoomOut := widgets.NewSDLImage(0, 0, btnHeight, btnHeight, BUTTON_ZOOM_IN, "zoomout", 0, 1, widgets.WIDGET_STYLE_DRAW_BORDER_AND_BG, 0, func(s string, b, i1, i2 int32) bool {
		scaleEverything(renderer, 0.9, widgetGroup)
		return true
	})

	labelGen := widgets.NewSDLLabel(0, 0, 600, btnHeight, LABEL_GEN, "Gen:0", widgets.ALIGN_LEFT, widgets.WIDGET_STYLE_NONE)
	labelSpeed := widgets.NewSDLLabel(0, 0, 350, btnHeight, LABEL_SPEED, "Delay:0ms", widgets.ALIGN_LEFT, widgets.WIDGET_STYLE_NONE)
	statusLabel = widgets.NewSDLLabel(0, viewport.H-btnHeight, viewport.W, btnHeight, LABEL_LOG, "Delay:0ms", widgets.ALIGN_LEFT, widgets.WIDGET_STYLE_DRAW_BORDER)
	statusLabel.SetBorderColour(&sdl.Color{R: 100, G: 100, B: 100, A: 255})

	var loadFile *widgets.SDL_Button
	var fileList *widgets.SDL_FileList
	pathEntry := widgets.NewSDLEntry(0, 0, 500, btnHeight, PATH_ENTRY, rleFile.Filename(), widgets.WIDGET_STYLE_DRAW_BORDER_AND_BG, func(old, new string, t widgets.ENTRY_EVENT_TYPE) (string, error) {
		switch t {
		case widgets.ENTRY_EVENT_FOCUS:
			fileList.Show(viewport)
		case widgets.ENTRY_EVENT_UN_FOCUS:
		}
		_, err := os.Stat(new)
		setErrorStatus(err)
		return new, err
	})

	fileList, err = widgets.NewFileList(0, 0, btnHeight, LIST_FILES, path.Dir(rleFile.Filename()), nil, widgets.WIDGET_STYLE_DRAW_BORDER, func(s string, i widgets.FILE_LIST_RESPONSE_CODE, id int32) bool {
		switch i {
		case widgets.FILE_LIST_FILE_SELECT:
			pathEntry.SetText(s)
			return true
		case widgets.FILE_LIST_PATH_SELECT:
			fileList.Reload(s)
			return true
		}
		return false
	}, func(b bool, s string) bool {
		if b {
			return !strings.HasPrefix(strings.ToLower(s), ".")
		}
		return strings.HasSuffix(strings.ToLower(s), ".rle")
	})

	fileList.SetVisible(false)
	fileList.SetLog(func(l widgets.LOG_LEVEL, s string) {
		setStatus(l, s)
	})

	loadFile = widgets.NewSDLButton(0, 0, btnWidth, btnHeight, BUTTON_LOAD_FILE, "Load", widgets.WIDGET_STYLE_DRAW_BORDER_AND_BG, 500, func(s string, b, i1, i2 int32) bool {
		err := loadRleFile(pathEntry.GetText())
		setErrorStatus(err)
		return true
	})

	btnStep := widgets.NewSDLButton(0, 0, btnWidth, btnHeight, BUTTON_STEP, "Step", widgets.WIDGET_STYLE_DRAW_BORDER_AND_BG, 0, func(s string, b, i1, i2 int32) bool {
		lifeGen.SetRunFor(1, nil)
		return true
	})

	var btnStop *widgets.SDL_Button
	btnStop = widgets.NewSDLButton(0, 0, btnWidth+30, btnHeight, BUTTON_STOP_START, "Stop", widgets.WIDGET_STYLE_DRAW_BORDER_AND_BG, 500, func(s string, b, i1, i2 int32) bool {
		if lifeGen.IsRunning() {
			lifeGen.SetRunFor(0, nil)
			btnStop.SetText("Start")
			mouseOn = true
		} else {
			lifeGen.SetRunFor(go_life.RUN_FOR_EVER, nil)
			btnStop.SetText("Stop")
			mouseOn = false
		}
		return true
	})

	arrowR := widgets.NewSDLShapeArrowRight(arrowPosX, arrowPosY, 100, 100, ARROW_RIGHT, widgets.WIDGET_STYLE_DRAW_BORDER_AND_BG, func(s string, b, i1, i2 int32) bool {
		cellOffsetX = cellOffsetX + 100
		return true
	})
	arrowD := widgets.NewSDLShapeArrowRight(arrowPosX, arrowPosY, 100, 100, ARROW_DOWN, widgets.WIDGET_STYLE_DRAW_BORDER_AND_BG, func(s string, b, i1, i2 int32) bool {
		cellOffsetY = cellOffsetY + 100
		return true
	})
	arrowD.Rotate(90)
	arrowL := widgets.NewSDLShapeArrowRight(arrowPosX, arrowPosY, 100, 100, ARROW_LEFT, widgets.WIDGET_STYLE_DRAW_BORDER_AND_BG, func(s string, b, i1, i2 int32) bool {
		cellOffsetX = cellOffsetX - 100
		return true
	})
	arrowL.Rotate(180)
	arrowU := widgets.NewSDLShapeArrowRight(arrowPosX, arrowPosY, 100, 100, ARROW_UP, widgets.WIDGET_STYLE_DRAW_BORDER_AND_BG, func(s string, b, i1, i2 int32) bool {
		cellOffsetY = cellOffsetY - 100
		return true
	})

	arrowU.Rotate(-90)

	buttonsTL.Add(btnClose)
	buttonsTL.Add(btnStop)
	buttonsTL.Add(labelGen)
	buttonsTL.Add(widgets.NewSDLSeparator(0, 0, 20, btnHeight, 999, widgets.WIDGET_STYLE_DRAW_BORDER))

	buttonsTR.Add(btnZoomIn)
	buttonsTR.Add(btnZoomOut)

	buttonsRunning.Add(btnSlower)
	buttonsRunning.Add(btnFaster)
	buttonsRunning.Add(btnFastest)
	buttonsRunning.Add(labelSpeed)
	buttonsRunning.Add(widgets.NewSDLSeparator(0, 0, 20, btnHeight, 999, widgets.WIDGET_STYLE_DRAW_BORDER))

	buttonsPaused.Add(btnStep)
	buttonsRunning.Add(widgets.NewSDLSeparator(0, 0, 20, btnHeight, 999, widgets.WIDGET_STYLE_DRAW_BORDER))
	buttonsRunning.Add(loadFile)
	buttonsRunning.Add(pathEntry)
	buttonsRunning.Add(fileList)

	arrows.Add(arrowR)
	arrows.Add(arrowD)
	arrows.Add(arrowL)
	arrows.Add(arrowU)

	statusGroup.Add(statusLabel)

	buttonsRunning.SetVisible(true)
	setErrorStatus(nil)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load 'lem': %s\n", err)
		return 1
	}

	go func() {
		for running {
			labelGen.SetText(fmt.Sprintf("Gen:%d(%d)", lifeGen.GetGenerationCount(), lifeGenSpeed))
			updateButtons(widgetGroup, renderer.GetViewport())
			sdl.Delay(200)
		}
	}()

	go func() {
		sdl.Delay(1000)
		for running {
			lifeGenTime = sdl.GetTicks64()
			lifeGen.NextGen()
			lifeGenSpeed = sdl.GetTicks64() - lifeGenTime
			if lifeGenSpeed < uint64(loopDelay) {
				sdl.Delay(uint32(loopDelay - lifeGenSpeed))
			}
		}
	}()

	for running {
		viewport = renderer.GetViewport()
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
				w := widgetGroup.InsideWidget(t.X, t.Y)
				if w != nil && w.GetWidgetId() == mouseData.GetWidgetId() && t.State == sdl.PRESSED {
					w.Click(mouseData.ActionStartDragging(t))
				} else {
					mouseData.ActionNotDragging(t)
				}
			case *sdl.MouseButtonEvent:
				w := widgetGroup.InsideWidget(t.X, t.Y)
				if w != nil {
					if t.State == sdl.PRESSED {
						widgetGroup.SetFocusedId(w.GetWidgetId())
						go w.Click(mouseData.ActionMouseDown(t, w.GetWidgetId()))
					} else {
						if mouseData.IsDragging() {
							w.Click(mouseData.ActionStopDragging(t))
							mouseData.ActionReset(t)
						}
					}
				} else {
					if widgetGroup.GetFocusedWidget() == nil {
						cellOffsetX, cellOffsetY = centerOnXY(t.X, t.Y, lifeGen)
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
		renderer.FillRect(&sdl.Rect{X: 0, Y: 0, W: viewport.W, H: btnTopMarginHeight})
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

func updateButtons(wg *widgets.SDL_WidgetGroup, view sdl.Rect) {
	var listPaused *widgets.SDL_WidgetSubGroup
	var listRunning *widgets.SDL_WidgetSubGroup
	var listTopLeft *widgets.SDL_WidgetSubGroup
	var fileList *widgets.SDL_FileList
	var pathEntry *widgets.SDL_Entry
	var buttonLoadFile *widgets.SDL_Button
	wl := wg.AllWidgets()
	for _, ww := range wl {
		switch ww.GetWidgetId() {
		case BUTTON_FASTEST, BUTTON_FASTER:
			ww.SetEnabled(loopDelay > MIN_LOOP_DELAY)
		case BUTTON_SLOWER:
			ww.SetEnabled(loopDelay < MAX_LOOP_DELAY)
		case LABEL_SPEED:
			ww.(*widgets.SDL_Label).SetText(fmt.Sprintf("Speed:%d", (MAX_LOOP_DELAY/50)-(loopDelay/50)))
		case PATH_ENTRY:
			pathEntry = ww.(*widgets.SDL_Entry)
		case LABEL_LOG:
			_, h := ww.GetSize()
			ww.SetPosition(0, view.H-h)
			ww.SetSize(view.W, -1)
		case LIST_FILES:
			fileList = ww.(*widgets.SDL_FileList)
		case BUTTON_LOAD_FILE:
			buttonLoadFile = ww.(*widgets.SDL_Button)
		}
	}

	ll := wg.AllSubGroups()
	for _, l := range ll {
		switch l.GetWidgetId() {
		case LIST_ARROWS:
			l.SetVisible(lifeGen.GetRunFor() < 2)
		case LIST_TOP_LEFT:
			listTopLeft = l
		case LIST_TOP_RIGHT:
			l.ArrangeRL(view.W-btnGap, btnMarginTopS, btnGap)
		case LIST_PAUSED:
			listPaused = l
		case LIST_RUNNING:
			listRunning = l
		}
	}

	x, y := listTopLeft.ArrangeLR(btnGap, btnMarginTopS, btnGap)
	listRunning.ArrangeLR(x, y, btnGap)
	listPaused.ArrangeLR(x, y, btnGap)
	listRunning.SetVisible(lifeGen.GetRunFor() >= 2)
	listPaused.SetVisible(lifeGen.GetRunFor() < 2)

	if pathEntry != nil && fileList != nil {
		x, y := pathEntry.GetPosition()
		pathEntry.SetSize(view.W-x-300, -1)
		fileList.SetPosition(pathEntry.GetRect().X, y+pathEntry.GetRect().H-2)
		x, y = fileList.GetPosition()
		v := &sdl.Rect{X: x, Y: y, W: viewport.W - x - 100, H: viewport.H - y}
		fileList.ArrangeGrid(v, 2, btnHeight, []int32{0})
		if buttonLoadFile != nil {
			_, err := os.Stat(pathEntry.GetText())
			pathEntry.SetError(err != nil)
			buttonLoadFile.SetEnabled(err == nil)
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

func setErrorStatus(e error) {
	if e == nil {
		statusGroup.SetVisible(false)
	} else {
		s := strings.TrimPrefix(e.Error(), "stat ")
		statusLabel.SetForeground(&sdl.Color{R: 255, G: 0, B: 0, A: 255})
		statusGroup.SetVisible(true)
		statusLabel.SetText(s)
	}
}

func setStatus(l widgets.LOG_LEVEL, s string) {
	if s == "" {
		statusGroup.SetVisible(false)
	} else {
		statusLabel.SetForeground(&sdl.Color{R: 255, G: 0, B: 0, A: 255})
		statusGroup.SetVisible(true)
		statusLabel.SetText(fmt.Sprintf("%d: %s", l, s))
	}
}

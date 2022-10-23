package main

import (
	"time"

	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

/****************************************************************************************
* SDL_Arrow code
* Implements SDL_Widget cos it is one!
**/
type SDL_Arrow struct {
	SDL_WidgetBase
	shape   *SDL_Shape
	onClick func(SDL_Widget, int32, int32) bool
}

var _ SDL_Widget = (*SDL_Arrow)(nil) // Ensure SDL_Button 'is a' SDL_Widget

func NewSDLArrow(x, y, w, h, id int32, bgColour, fgColour *sdl.Color, deBounce int, onClick func(SDL_Widget, int32, int32) bool) *SDL_Arrow {
	but := &SDL_Arrow{onClick: onClick}
	but.SDL_WidgetBase = initBase(x, y, w, h, id, deBounce, bgColour, fgColour)
	but.defineShape()
	return but
}

// func (b *SDL_Arrow) GetRect() *sdl.Rect {
// 	return b.rect
// }

func (b *SDL_Arrow) defineShape() {
	ww := b.w
	if ww < 0 {
		ww = ww * -1
	}
	hh := b.h
	if hh < 0 {
		hh = hh * -1
	}
	w := b.w
	h := b.h
	x := b.x
	y := b.y

	sh := NewSDLShape()
	if ww > hh {
		var halfH int32 = h / 2
		var qtr1H int32 = h / 4
		var thrd1W int32 = w / 6
		var thrd2W int32 = thrd1W * 4
		sh.Add(x+thrd1W, y-qtr1H)
		sh.Add(x+thrd2W, y-qtr1H)
		sh.Add(x+thrd2W, y-halfH)
		sh.Add(x+w, y)
		sh.Add(x+thrd2W, y+halfH)
		sh.Add(x+thrd2W, y+qtr1H)
		sh.Add(x+thrd1W, y+qtr1H)
	} else {
		var halfW int32 = b.w / 2
		var qtr1W int32 = b.w / 4
		var thrd1H int32 = b.h / 6
		var thrd2H int32 = thrd1H * 4
		sh.Add(x+qtr1W, y+thrd1H)
		sh.Add(x+qtr1W, y+thrd2H)
		sh.Add(x+halfW, y+thrd2H)
		sh.Add(x, y+h)
		sh.Add(x-halfW, y+thrd2H)
		sh.Add(x-qtr1W, y+thrd2H)
		sh.Add(x-qtr1W, y+thrd1H)
	}
	r := sh.Rect()
	b.SetPosition(r.X, r.Y)
	b.SetSize(r.W, r.H)
	b.shape = sh
}

// func (b *SDL_Arrow) updatePositionSize() {
// 	var minx int16 = math.MaxInt16
// 	var miny int16 = math.MaxInt16
// 	var maxx int16 = math.MinInt16
// 	var maxy int16 = math.MinInt16
// 	vx := b.vx
// 	vy := b.vy

// 	for i := 0; i < len(vx); i++ {
// 		if vx[i] < minx {
// 			minx = vx[i]
// 		}
// 		if vx[i] > maxx {
// 			maxx = vx[i]
// 		}
// 		if vy[i] < miny {
// 			miny = vy[i]
// 		}
// 		if vy[i] > maxy {
// 			maxy = vy[i]
// 		}
// 	}
// 	b.SetPosition(int32(minx), int32(miny))
// 	b.SetSize(int32(maxx-minx), int32(maxy-miny))
// }

func (b *SDL_Arrow) SetOnClick(f func(SDL_Widget, int32, int32) bool) {
	b.onClick = f
}

func (b *SDL_Arrow) Click(x, y int32) bool {
	if b.enabled && b.visible && b.notPressed && b.onClick != nil {
		if b.deBounce > 0 {
			b.notPressed = false
			defer func() {
				time.Sleep(time.Millisecond * time.Duration(b.deBounce))
				b.notPressed = true
			}()
		}
		return b.onClick(b, x, y)
	}
	return false
}

func (b *SDL_Arrow) Destroy() {

}

func (b *SDL_Arrow) Draw(renderer *sdl.Renderer, font *ttf.Font) error {
	if b.visible {
		renderer.SetDrawColor(b.bg.R, b.bg.G, b.bg.B, b.bg.A)
		gfx.FilledPolygonColor(renderer, b.shape.vx, b.shape.vy, *widgetColourDim(b.bg, b.IsEnabled(), 2))
		gfx.PolygonColor(renderer, b.shape.vx, b.shape.vy, *widgetColourDim(b.fg, b.IsEnabled(), 2))
	}
	return nil
}

/****************************************************************************************
* SDL_Button code
* Implements SDL_Widget cos it is one!
* Implements SDL_TextWidget because it has text and uses the texture cache
**/
type SDL_Button struct {
	SDL_WidgetBase
	text         string
	textureCache *SDL_TextureCache
	onClick      func(SDL_Widget, int32, int32) bool
}

var _ SDL_Widget = (*SDL_Button)(nil)     // Ensure SDL_Button 'is a' SDL_Widget
var _ SDL_TextWidget = (*SDL_Button)(nil) // Ensure SDL_Button 'is a' SDL_TextWidget

func NewSDLButton(x, y, w, h, id int32, text string, bgColour, fgColour *sdl.Color, deBounce int, onClick func(SDL_Widget, int32, int32) bool) *SDL_Button {
	but := &SDL_Button{text: text, onClick: onClick}
	but.SDL_WidgetBase = initBase(x, y, w, h, id, deBounce, bgColour, fgColour)
	return but
}

func (b *SDL_Button) SetOnClick(f func(SDL_Widget, int32, int32) bool) {
	b.onClick = f
}

func (b *SDL_Button) SetText(text string) {
	b.text = text
}

func (b *SDL_Button) SetTextureCache(tc *SDL_TextureCache) {
	b.textureCache = tc
}

func (b *SDL_Button) GetTextureCache() *SDL_TextureCache {
	return b.textureCache
}

func (b *SDL_Button) GetText() string {
	return b.text
}

func (b *SDL_Button) Click(x, y int32) bool {
	if b.enabled && b.visible && b.notPressed && b.onClick != nil {
		if b.deBounce > 0 {
			b.notPressed = false
			defer func() {
				time.Sleep(time.Millisecond * time.Duration(b.deBounce))
				b.notPressed = true
			}()
		}
		return b.onClick(b, x, y)
	}
	return false
}

func (b *SDL_Button) Destroy() {

}

func (b *SDL_Button) Draw(renderer *sdl.Renderer, font *ttf.Font) error {
	if b.visible {
		// Save the current Draw colour and restore on exit
		rr, gg, bb, aa, _ := renderer.GetDrawColor()
		defer renderer.SetDrawColor(rr, gg, bb, aa)

		// Dray the button
		renderer.SetDrawColor(b.bg.R, b.bg.G, b.bg.B, b.bg.A)
		renderer.FillRect(&sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
		borderColour := widgetColourBright(b.bg, !b.enabled)
		renderer.SetDrawColor(borderColour.R, borderColour.G, borderColour.B, borderColour.A)
		renderer.DrawRect(&sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
		renderer.DrawRect(&sdl.Rect{X: b.x + 1, Y: b.y + 1, W: b.w - 2, H: b.h - 2})

		// get data from the cache entry
		ctwe, err := getCachedTextWidgetEntry(renderer, b, font)
		if err != nil {
			return err
		}
		clip := ctwe.clipRect
		// Center the text inside the buttonj
		tx := b.x + (b.w-clip.W)/2
		ty := b.y + (b.h-clip.H)/2
		renderer.Copy(ctwe.texture, nil, &sdl.Rect{X: tx, Y: ty, W: clip.W, H: clip.H})
	}
	return nil
}

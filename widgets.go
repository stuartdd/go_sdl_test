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
	vx      []int16
	vy      []int16
	onClick func(SDL_Widget, int32, int32) bool
}

var _ SDL_Widget = (*SDL_Arrow)(nil) // Ensure SDL_Button 'is a' SDL_Widget

func NewSDLArrow(x, y, w, h, dir int32, bgColour, fgColour *sdl.Color, deBounce int, onClick func(SDL_Widget, int32, int32) bool) *SDL_Arrow {
	but := &SDL_Arrow{onClick: onClick}
	but.SDL_WidgetBase = initBase(x, y, w, h, deBounce, bgColour, fgColour)
	but.shape()
	return but
}

func (b *SDL_Arrow) shape() {
	b.vx = make([]int16, 8)
	b.vy = make([]int16, 8)

	w := b.w
	if w < 0 {
		w = w * -1
	}
	h := b.h
	if h < 0 {
		h = h * -1
	}

	b.vx[0] = int16(b.x)
	b.vy[0] = int16(b.y)

	if w > h {
		var halfH int32 = b.h / 2
		var qtr1H int32 = b.h / 4
		var thrd1W int32 = b.w / 3
		var thrd2W int32 = thrd1W * 2
		b.vx[1] = int16(b.x + thrd1W)
		b.vy[1] = int16(b.y - qtr1H)
		b.vx[2] = int16(b.x + thrd2W)
		b.vy[2] = int16(b.y - qtr1H)
		b.vx[3] = int16(b.x + thrd2W)
		b.vy[3] = int16(b.y - halfH)
		b.vx[4] = int16(b.x + b.w)
		b.vy[4] = int16(b.y)
		b.vx[5] = int16(b.x + thrd2W)
		b.vy[5] = int16(b.y + halfH)
		b.vx[6] = int16(b.x + thrd2W)
		b.vy[6] = int16(b.y + qtr1H)
		b.vx[7] = int16(b.x + thrd1W)
		b.vy[7] = int16(b.y + qtr1H)
	} else {
		var halfW int32 = b.w / 2
		var qtr1W int32 = b.w / 4
		var thrd1H int32 = b.h / 3
		var thrd2H int32 = thrd1H * 2
		b.vx[1] = int16(b.x + qtr1W)
		b.vy[1] = int16(b.y + thrd1H)
		b.vx[2] = int16(b.x + qtr1W)
		b.vy[2] = int16(b.y + thrd2H)
		b.vx[3] = int16(b.x + halfW)
		b.vy[3] = int16(b.y + thrd2H)
		b.vx[4] = int16(b.x)
		b.vy[4] = int16(b.y + b.h)
		b.vx[5] = int16(b.x - halfW)
		b.vy[5] = int16(b.y + thrd2H)
		b.vx[6] = int16(b.x - qtr1W)
		b.vy[6] = int16(b.y + thrd2H)
		b.vx[7] = int16(b.x - qtr1W)
		b.vy[7] = int16(b.y + thrd1H)
	}
}

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
		gfx.FilledPolygonColor(renderer, b.vx, b.vy, *widgetColourDim(b.bg, b.IsEnabled(), 2))
		gfx.PolygonColor(renderer, b.vx, b.vy, *widgetColourDim(b.fg, b.IsEnabled(), 2))
		gfx.FilledCircleColor(renderer, b.x, b.y, 5, sdl.Color{R: 255, G: 0, B: 0, A: 255})
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

func NewSDLButton(x, y, w, h int32, text string, bgColour, fgColour *sdl.Color, deBounce int, onClick func(SDL_Widget, int32, int32) bool) *SDL_Button {
	but := &SDL_Button{text: text, onClick: onClick}
	but.SDL_WidgetBase = initBase(x, y, w, h, deBounce, bgColour, fgColour)
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

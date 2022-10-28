package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type SDL_Label struct {
	SDL_WidgetBase
	text         string
	cacheKey     string
	cacheInvalid bool
	textureCache *SDL_TextureCache
	align        ALIGN_TEXT
}

var _ SDL_TextWidget = (*SDL_Label)(nil)         // Ensure SDL_Button 'is a' SDL_Widget
var _ SDL_Widget = (*SDL_Label)(nil)             // Ensure SDL_Button 'is a' SDL_Widget
var _ SDL_TextureCacheWidget = (*SDL_Label)(nil) // Ensure SDL_Button 'is a' SDL_Widget

func NewSDLLabel(x, y, w, h, id int32, text string, align ALIGN_TEXT, bgColour, fgColour *sdl.Color) *SDL_Label {
	but := &SDL_Label{text: text, cacheInvalid: true, align: align, cacheKey: fmt.Sprintf("label:%d:%d", id, rand.Intn(100))}
	but.SDL_WidgetBase = initBase(x, y, w, h, id, 0, bgColour, fgColour)
	return but
}

func (b *SDL_Label) SetTextureCache(tc *SDL_TextureCache) {
	b.textureCache = tc
}

func (b *SDL_Label) GetTextureCache() *SDL_TextureCache {
	return b.textureCache
}

func (b *SDL_Label) SetForeground(c *sdl.Color) {
	if b.fg != c {
		b.cacheInvalid = true
		b.fg = c
	}
}

func (b *SDL_Label) SetText(text string) {
	if b.text != text {
		b.cacheInvalid = true
		b.text = text
	}
}

func (b *SDL_Label) SetEnabled(e bool) {
	if b.IsEnabled() != e {
		b.cacheInvalid = true
		b.SDL_WidgetBase.SetEnabled(e)
	}
}

func (b *SDL_Label) GetText() string {
	return b.text
}

func (b *SDL_Label) Click(x, y int32) bool {
	return false
}

func (b *SDL_Label) Draw(renderer *sdl.Renderer, font *ttf.Font) error {
	if b.IsVisible() {
		ctwe, ok := b.textureCache.textureMap[b.cacheKey]
		if !ok || b.cacheInvalid {
			var err error
			ctwe, err = getTextureCacheEntryForText(renderer, b.text, b.cacheKey, font, widgetColourDim(b.fg, b.IsEnabled(), 2))
			if err != nil {
				renderer.SetDrawColor(255, 0, 0, 255)
				renderer.DrawRect(&sdl.Rect{X: b.x, Y: b.y, W: 100, H: 100})
				return nil
			}
			b.textureCache.textureMap[b.cacheKey] = ctwe
			if b.align == ALIGN_FIT {
				b.SetSize(ctwe.clipRect.W, b.h)
			}
		}
		clip := ctwe.clipRect
		// Center the text inside the buttonj
		var tx int32
		switch b.align {
		case ALIGN_CENTER:
			tx = b.x + (b.w-clip.W)/2
		case ALIGN_LEFT:
			tx = b.x + 10
		case ALIGN_RIGHT:
			tx = b.x + (b.w - clip.W)
		case ALIGN_FIT:
			tx = b.x + 10
			if b.align == ALIGN_FIT {
				b.SetSize(ctwe.clipRect.W+20, b.h)
			}
		}
		if b.align == ALIGN_CENTER {
			tx = b.x + (b.w-clip.W)/2
		}
		ty := b.y + (b.h-clip.H)/2
		fullRect := &sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h}
		if b.bg != nil {
			renderer.SetDrawColor(b.bg.R, b.bg.G, b.bg.B, b.bg.A)
			renderer.FillRect(fullRect)
		}
		renderer.Copy(ctwe.texture, nil, &sdl.Rect{X: tx, Y: ty, W: clip.W, H: clip.H})
		if b.fg != nil {
			renderer.SetDrawColor(b.fg.R, b.fg.G, b.fg.B, b.fg.A)
			renderer.DrawRect(fullRect)
		}
	}
	return nil
}
func (b *SDL_Label) Destroy() {
	// Image cache takes care of all images!
}

/****************************************************************************************
* SDL_Image code
* Implements SDL_Widget cos it is one!
* Implements SDL_TextWidget because it has text and uses the texture cache
**/
type SDL_Separator struct {
	SDL_WidgetBase
}

var _ SDL_Widget = (*SDL_Separator)(nil) // Ensure SDL_Button 'is a' SDL_Widget

func NewSDLSeparator(x, y, w, h, id int32, bgColour *sdl.Color) *SDL_Separator {
	but := &SDL_Separator{}
	but.SDL_WidgetBase = initBase(x, y, w, h, id, 0, bgColour, nil)
	return but
}

func (b *SDL_Separator) Click(x, y int32) bool {
	return false
}

func (b *SDL_Separator) Draw(renderer *sdl.Renderer, font *ttf.Font) error {
	if b.bg != nil {
		renderer.SetDrawColor(b.bg.R, b.bg.G, b.bg.B, b.bg.A)
		renderer.FillRect(&sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
	}
	return nil
}

func (b *SDL_Separator) Destroy() {
	// Image cache takes care of all images!
}

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
* SDL_Image code
* Implements SDL_Widget cos it is one!
* Implements SDL_TextWidget because it has text and uses the texture cache
**/
type SDL_Image struct {
	SDL_WidgetBase
	textureName  string
	frame        int32
	frameCount   int32
	textureCache *SDL_TextureCache
	onClick      func(SDL_Widget, int32, int32) bool
}

var _ SDL_Widget = (*SDL_Image)(nil)      // Ensure SDL_Image 'is a' SDL_Widget
var _ SDL_ImageWidget = (*SDL_Image)(nil) // Ensure SDL_Image 'is a' SDL_ImageWidget

func NewSDLImage(x, y, w, h, id int32, textureName string, frame, frameCount int32, bgColour, fgColour *sdl.Color, deBounce int, onClick func(SDL_Widget, int32, int32) bool) *SDL_Image {
	but := &SDL_Image{textureName: textureName, frame: frame, frameCount: frameCount, onClick: onClick}
	but.SDL_WidgetBase = initBase(x, y, w, h, id, deBounce, bgColour, fgColour)
	return but
}

func (b *SDL_Image) SetFrame(tf int32) {
	if tf >= b.frameCount {
		tf = 0
	}
	b.frame = tf
}

func (b *SDL_Image) GetFrame() int32 {
	return b.frame
}

func (b *SDL_Image) NextFrame() int32 {
	b.frame++
	if b.frame >= b.frameCount {
		b.frame = 0
	}
	return b.frame
}

func (b *SDL_Image) GetFrameCount() int32 {
	return b.frameCount
}

func (b *SDL_Image) SetTextureCache(tc *SDL_TextureCache) {
	b.textureCache = tc
}

func (b *SDL_Image) GetTextureCache() *SDL_TextureCache {
	return b.textureCache
}

func (b *SDL_Image) Click(x, y int32) bool {
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

func (b *SDL_Image) Draw(renderer *sdl.Renderer, font *ttf.Font) error {
	if b.visible {
		borderRect := &sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h}
		outRect := &sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h}
		var bg *sdl.Color = nil
		var fg *sdl.Color = nil
		if b.IsEnabled() {
			fg = b.fg
			bg = b.bg
		} else {
			outRect = widgetShrinkRect(outRect, 4)
			fg = widgetColourDim(b.fg, false, 2)
		}
		if bg != nil {
			// Background
			renderer.SetDrawColor(b.bg.R, b.bg.G, b.bg.B, b.bg.A)
			renderer.FillRect(borderRect)
		}
		image, ir, err := b.textureCache.GetTexture(b.textureName)
		if err != nil {
			renderer.SetDrawColor(255, 0, 0, 255)
			renderer.DrawRect(&sdl.Rect{X: b.x, Y: b.y, W: 100, H: 100})
			return nil
		}
		if bg != nil || fg != nil {
			outRect = widgetShrinkRect(outRect, 8)
		}
		if b.frameCount > 1 {
			w := (ir.W / b.frameCount)
			x := (w * b.frame)
			inRect := &sdl.Rect{X: x, Y: 0, W: w, H: outRect.H}
			outRect := &sdl.Rect{X: outRect.X, Y: outRect.Y, W: w, H: outRect.H}
			renderer.Copy(image, inRect, outRect)
		} else {
			renderer.Copy(image, nil, outRect)
		}
		// Border
		if fg != nil {
			renderer.SetDrawColor(fg.R, fg.G, fg.B, fg.A)
			renderer.DrawRect(borderRect)
		}
	}
	return nil
}

func (b *SDL_Image) Destroy() {
	// Image cache takes care of all images!
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

func (b *SDL_Button) GetText() string {
	return b.text
}

func (b *SDL_Button) SetTextureCache(tc *SDL_TextureCache) {
	b.textureCache = tc
}

func (b *SDL_Button) GetTextureCache() *SDL_TextureCache {
	return b.textureCache
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
		// Dray the button
		renderer.SetDrawColor(b.bg.R, b.bg.G, b.bg.B, b.bg.A)
		renderer.FillRect(&sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
		borderColour := widgetColourBright(b.bg, !b.enabled)
		renderer.SetDrawColor(borderColour.R, borderColour.G, borderColour.B, borderColour.A)
		renderer.DrawRect(&sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
		renderer.DrawRect(&sdl.Rect{X: b.x + 1, Y: b.y + 1, W: b.w - 2, H: b.h - 2})

		// get/create data from the cache entry
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

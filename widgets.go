package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

/****************************************************************************************
* SDL_Shape code.
* Implements SDL_Widget cos it is one!
**/
type SDL_Shape struct {
	SDL_WidgetBase
	validRect *sdl.Rect // If the state of the shape has changed this should be nil.
	vxIn      []int16
	vyIn      []int16
	vxOut     []int16
	vyOut     []int16
	onClick   func(SDL_Widget, int32, int32) bool
}

var _ SDL_Widget = (*SDL_Shape)(nil) // Ensure SDL_Button 'is a' SDL_Widget

func NewSDLShape(x, y, w, h, id int32, bgColour, fgColour *sdl.Color, onClick func(SDL_Widget, int32, int32) bool) *SDL_Shape {
	shape := &SDL_Shape{vxIn: make([]int16, 0), vyIn: make([]int16, 0), validRect: nil, onClick: onClick}
	shape.SDL_WidgetBase = initBase(x, y, w, h, id, 0, bgColour, fgColour)
	return shape
}

func NewSDLShapeArrowRight(x, y, w, h, id int32, bgColour, fgColour *sdl.Color, onClick func(SDL_Widget, int32, int32) bool) *SDL_Shape {
	sh := NewSDLShape(x, y, w, h, id, bgColour, fgColour, onClick)
	var halfH int32 = h / 2
	var qtr1H int32 = h / 4
	var thrd1W int32 = w / 6
	var thrd2W int32 = thrd1W * 4
	sh.Add(thrd1W, -qtr1H)
	sh.Add(thrd2W, -qtr1H)
	sh.Add(thrd2W, -halfH)
	sh.Add(w, 0)
	sh.Add(thrd2W, +halfH)
	sh.Add(thrd2W, +qtr1H)
	sh.Add(thrd1W, +qtr1H)
	return sh
}

func (s *SDL_Shape) SetPosition(x, y int32) bool {
	b := s.SDL_WidgetBase.SetPosition(x, y)
	if b {
		s.validRect = nil
	}
	return b
}

func (s *SDL_Shape) SetSize(w, h int32) bool {
	b := s.SDL_WidgetBase.SetSize(w, h)
	if b {
		s.validRect = nil
	}
	return b
}

func (s *SDL_Shape) Scale(sc float32) {
	s.SDL_WidgetBase.Scale(sc)
	for i := 0; i < len(s.vxIn); i++ {
		s.vxIn[i] = int16(float32(s.vxIn[i]) * sc)
		s.vyIn[i] = int16(float32(s.vyIn[i]) * sc)
	}
	s.validRect = nil
}

func (s *SDL_Shape) Add(x, y int32) {
	s.vxIn = append(s.vxIn, int16(x))
	s.vyIn = append(s.vyIn, int16(y))
	s.validRect = nil
}

func (b *SDL_Shape) Click(x, y int32) bool {
	if b.IsEnabled() && b.onClick != nil {
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

func (b *SDL_Shape) Destroy() {
}

func (s *SDL_Shape) Draw(renderer *sdl.Renderer, font *ttf.Font) error {
	if s.IsVisible() {
		s.GetRect() // Make sure we update the Out Arrays is the state of the shape was changed
		if s.bg != nil {
			gfx.FilledPolygonColor(renderer, s.vxOut, s.vyOut, *widgetColourDim(s.bg, s.IsEnabled(), 1.5))
		}
		if s.fg != nil {
			gfx.PolygonColor(renderer, s.vxOut, s.vyOut, *widgetColourDim(s.fg, s.IsEnabled(), 2.5))
		}
	}
	return nil
}

func (s *SDL_Shape) Rotate(angle int) {
	rad := float64(angle) * DEG_TO_RAD
	var px float64 = 0
	var py float64 = 0
	sinA := math.Sin(rad)
	cosA := math.Cos(rad)
	for i := 0; i < len(s.vxIn); i++ {
		px = float64(s.vxIn[i])
		py = float64(s.vyIn[i])
		s.vxIn[i] = int16(cosA*px - sinA*py)
		s.vyIn[i] = int16(sinA*px + cosA*py)
	}
	s.validRect = nil
}

func (s *SDL_Shape) Inside(x, y int32) bool {
	if s.visible {
		return isInsideRect(x, y, s.GetRect())
	}
	return false
}

func (s *SDL_Shape) GetRect() *sdl.Rect {
	if s.validRect == nil {
		count := len(s.vxIn)
		vxOut := make([]int16, count)
		vyOut := make([]int16, count)
		x := int16(s.x)
		y := int16(s.y)
		for i := 0; i < count; i++ {
			vxOut[i] = x + s.vxIn[i]
			vyOut[i] = y + s.vyIn[i]
		}

		var minx int16 = math.MaxInt16
		var miny int16 = math.MaxInt16
		var maxx int16 = math.MinInt16
		var maxy int16 = math.MinInt16
		for i := 0; i < count; i++ {
			if vxOut[i] < minx {
				minx = vxOut[i]
			}
			if vxOut[i] > maxx {
				maxx = vxOut[i]
			}
			if vyOut[i] < miny {
				miny = vyOut[i]
			}
			if vyOut[i] > maxy {
				maxy = vyOut[i]
			}
		}
		s.vxOut = vxOut
		s.vyOut = vyOut
		s.validRect = &sdl.Rect{X: int32(minx), Y: int32(miny), W: int32(maxx - minx), H: int32(maxy - miny)}
	}
	return s.validRect
}

/****************************************************************************************
* SDL_Label code
* Implements SDL_Widget cos it is one!
* Implements SDL_TextWidget because it has text and uses the texture cache
**/
type SDL_Label struct {
	SDL_WidgetBase
	text         string
	cacheKey     string
	cacheInvalid bool
	textureCache *SDL_TextureCache
	align        ALIGN_TEXT
}

var _ SDL_Widget = (*SDL_Label)(nil)             // Ensure SDL_Button 'is a' SDL_Widget
var _ SDL_TextWidget = (*SDL_Label)(nil)         // Ensure SDL_Button 'is a' SDL_TextWidget
var _ SDL_TextureCacheWidget = (*SDL_Label)(nil) // Ensure SDL_Button 'is a' SDL_TextureCacheWidget

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
		ctwe, ok := b.textureCache.Get(b.cacheKey)
		if !ok || b.cacheInvalid {
			var err error
			ctwe, err = NewTextureCacheEntryForString(renderer, b.text, b.cacheKey, font, widgetColourDim(b.fg, b.IsEnabled(), 2))
			if err != nil {
				renderer.SetDrawColor(255, 0, 0, 255)
				renderer.DrawRect(&sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
				return nil
			}
			b.textureCache.Add(b.cacheKey, ctwe)
			if b.align == ALIGN_FIT {
				b.SetSize(ctwe.W, b.h)
			}
		}
		aspect := float32(b.w) / float32(b.h)
		inset := float32(b.h) / 4
		th := float32(b.h) - inset
		tw := th * aspect
		var tx float32
		switch b.align {
		case ALIGN_CENTER:
			tx = (float32(b.w) - tw) / 2
		case ALIGN_LEFT:
			tx = 10
		case ALIGN_RIGHT:
			tx = float32(b.x+b.w) - tw
		}
		ty := (float32(b.h) - th) / 2

		if b.bg != nil {
			renderer.SetDrawColor(b.bg.R, b.bg.G, b.bg.B, b.bg.A)
			renderer.FillRect(&sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
		}
		renderer.Copy(ctwe.Texture, nil, &sdl.Rect{X: b.x + int32(tx), Y: b.y + int32(ty), W: int32(tw), H: int32(th)})
		if b.fg != nil {
			borderColour := widgetColourDim(b.fg, b.IsEnabled(), 2)
			renderer.SetDrawColor(borderColour.R, borderColour.G, borderColour.B, borderColour.A)
			renderer.DrawRect(&sdl.Rect{X: b.x + 1, Y: b.y + 1, W: b.w - 1, H: b.h - 1})
		}
	}
	return nil
}
func (b *SDL_Label) Destroy() {
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
	if b.IsEnabled() && b.onClick != nil {
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
		var cacheKey = fmt.Sprintf("%s.%s.%t", TEXTURE_CACHE_TEXT_PREF, b.text, b.IsEnabled() && b.notPressed)
		ctwe, ok := b.textureCache.Get(cacheKey)
		if !ok {
			var err error
			ctwe, err = NewTextureCacheEntryForString(renderer, b.text, cacheKey, font, widgetColourDim(b.fg, b.IsEnabled(), 2))
			if err != nil {
				renderer.SetDrawColor(255, 0, 0, 255)
				renderer.DrawRect(&sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
				return nil
			}
			b.textureCache.Add(cacheKey, ctwe)
		}
		if b.bg != nil {
			renderer.SetDrawColor(b.bg.R, b.bg.G, b.bg.B, b.bg.A)
			renderer.FillRect(&sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
		}
		// Center the text inside the buttonj
		aspect := float32(b.w) / float32(b.h)
		inset := float32(b.h) / 4
		th := float32(b.h) - inset
		tw := th * aspect
		tx := (float32(b.w) - tw) / 2
		ty := (float32(b.h) - th) / 2
		renderer.Copy(ctwe.Texture, nil, &sdl.Rect{X: b.x + int32(tx), Y: b.y + int32(ty), W: int32(tw), H: int32(th)})
		if b.fg != nil {
			borderColour := widgetColourDim(b.fg, b.IsEnabled(), 2)
			renderer.SetDrawColor(borderColour.R, borderColour.G, borderColour.B, borderColour.A)
			renderer.DrawRect(&sdl.Rect{X: b.x + 1, Y: b.y + 1, W: b.w - 1, H: b.h - 1})
		}
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
	if b.IsEnabled() && b.onClick != nil {
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
			fg = widgetColourDim(b.fg, false, 1.5)
		}
		if bg != nil {
			// Background
			renderer.SetDrawColor(b.bg.R, b.bg.G, b.bg.B, b.bg.A)
			renderer.FillRect(borderRect)
		}
		image, irw, _, err := b.textureCache.GetTextureForName(b.textureName)
		if err != nil {
			renderer.SetDrawColor(255, 0, 0, 255)
			renderer.DrawRect(&sdl.Rect{X: b.x, Y: b.y, W: 100, H: 100})
			return nil
		}
		if bg != nil || fg != nil {
			outRect = widgetShrinkRect(outRect, 4)
		}
		if b.frameCount > 1 {
			w := (irw / b.frameCount)
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
			renderer.DrawRect(&sdl.Rect{X: b.x + 1, Y: b.y + 1, W: b.w - 1, H: b.h - 1})
		}
	}
	return nil
}

func (b *SDL_Image) Destroy() {
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

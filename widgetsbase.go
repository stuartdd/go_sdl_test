package main

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type SDL_Widget interface {
	Draw(*sdl.Renderer, *ttf.Font) error
	Inside(int32, int32) bool
	Click(int32, int32) bool
	SetVisible(bool)
	IsVisible() bool
	SetEnabled(bool)
	IsEnabled() bool
	SetPosition(int32, int32)
	GetPosition() (int32, int32)
	SetSize(int32, int32)
	GetSize() (int32, int32)
	Destroy()
}

type SDL_TextWidget interface {
	SetTextureCache(*SDL_TextureCache)
	GetTextureCache() *SDL_TextureCache
	SetText(text string)
	GetText() string
	IsEnabled() bool
	GetForeground() *sdl.Color
}

type SDL_WidgetBase struct {
	x, y, w, h int32
	visible    bool
	enabled    bool
	notPressed bool
	deBounce   int
	bg         *sdl.Color
	fg         *sdl.Color
}

/****************************************************************************************
* Common (base) functions for ALL SDL_Widget instances
**/
func initBase(x, y, w, h int32, deBounce int, bgColour, fgColour *sdl.Color) SDL_WidgetBase {
	return SDL_WidgetBase{x: x, y: y, w: w, h: h, enabled: true, visible: true, notPressed: true, deBounce: deBounce, bg: bgColour, fg: fgColour}
}

func (b *SDL_WidgetBase) SetPosition(x, y int32) {
	b.x = x
	b.y = y
}

func (b *SDL_WidgetBase) GetPosition() (int32, int32) {
	return b.x, b.y
}

func (b *SDL_WidgetBase) SetSize(w, h int32) {
	b.w = w
	b.h = h
}

func (b *SDL_WidgetBase) GetSize() (int32, int32) {
	return b.w, b.h
}

func (b *SDL_WidgetBase) SetVisible(v bool) {
	b.visible = v
}

func (b *SDL_WidgetBase) IsVisible() bool {
	return b.visible
}

func (b *SDL_WidgetBase) SetEnabled(e bool) {
	b.enabled = e
}

func (b *SDL_WidgetBase) IsEnabled() bool {
	return b.enabled && b.notPressed
}

func (b *SDL_WidgetBase) SetDeBounce(db int) {
	b.deBounce = db
}

func (b *SDL_WidgetBase) GetDebounce() int {
	return b.deBounce
}

func (b *SDL_WidgetBase) GetForeground() *sdl.Color {
	return b.fg
}

func (b *SDL_WidgetBase) GetBackground() *sdl.Color {
	return b.bg
}

func (b *SDL_WidgetBase) Inside(x, y int32) bool {
	if b.visible {
		if x < b.x {
			return false
		}
		if y < b.y {
			return false
		}
		if x > (b.x + b.w) {
			return false
		}
		if y > (b.y + b.h) {
			return false
		}
		return true
	}
	return false
}

/****************************************************************************************
* Container for SDL_Widget instances.
**/
type SDL_Widgets struct {
	font         *ttf.Font
	list         []SDL_Widget
	textureCache *SDL_TextureCache
}

func NewSDLWidgets(font *ttf.Font) *SDL_Widgets {
	return &SDL_Widgets{textureCache: NewTextureCache(), list: make([]SDL_Widget, 0), font: font}
}

func (wl *SDL_Widgets) Add(widget SDL_Widget) {
	tw, ok := widget.(SDL_TextWidget)
	if ok {
		tw.SetTextureCache(wl.textureCache)
	}
	wl.list = append(wl.list, widget)
}

func (wl *SDL_Widgets) Inside(x, y int32) SDL_Widget {
	for _, w := range wl.list {
		if w.Inside(x, y) {
			return w
		}
	}
	return nil
}

func (wl *SDL_Widgets) Draw(renderer *sdl.Renderer) {
	for _, w := range wl.list {
		w.Draw(renderer, wl.font)
	}
}

func (wl *SDL_Widgets) SetFont(font *ttf.Font) {
	wl.font = font
}

func (wl *SDL_Widgets) GetFont() *ttf.Font {
	return wl.font
}

func (wl *SDL_Widgets) Destroy() {
	for _, w := range wl.list {
		w.Destroy()
	}
	if wl.textureCache != nil {
		wl.textureCache.Destroy()
	}
}

/****************************************************************************************
* Texture cache for widgets that have text to display.
* Textures are sdl resources and need to be Destroyed.
* SDL_Widgets destroys all textures via the SDL_Widgets Destroy() function.
*
* USE:	widigts := NewSDLWidgets(font)
* 		defer widigts.Destroy()
**/

type SDL_TextureCacheEntry struct {
	clipRect sdl.Rect
	texture  *sdl.Texture
}

func (tce *SDL_TextureCacheEntry) Destroy() {
	if tce.texture != nil {
		tce.texture.Destroy()
	}
}

type SDL_TextureCache struct {
	textureMap map[string]*SDL_TextureCacheEntry
}

func NewTextureCache() *SDL_TextureCache {
	return &SDL_TextureCache{textureMap: make(map[string]*SDL_TextureCacheEntry)}
}

func (wl *SDL_TextureCache) Destroy() {
	for _, v := range wl.textureMap {
		v.Destroy()
	}
}

/****************************************************************************************
* Utilities
* getCachedTextWidgetEntry Returns cached texture data
*
* widgetColourDim takes a colour and returns a dimmer by same colour. Used for disabled widget text
* widgetColourBright takes a colour and returns a brighter by same colour. Used for Widget Borders
*
**/
func getCachedTextWidgetEntry(renderer *sdl.Renderer, tw SDL_TextWidget, font *ttf.Font) (*SDL_TextureCacheEntry, error) {
	// cache key must include Dim variations of the textures
	var cacheKey = fmt.Sprintf("%s%t", tw.GetText(), tw.IsEnabled())
	var cacheDataEntry *SDL_TextureCacheEntry
	tc := tw.GetTextureCache()

	// Do we have a texture cache and is the texture already cached
	if tc != nil {
		cacheDataEntry = tc.textureMap[cacheKey]
	}
	// If not cached than creat a cache entry
	if cacheDataEntry == nil {
		text, err := font.RenderUTF8Blended(tw.GetText(), *widgetColourDim(tw.GetForeground(), tw.IsEnabled(), 3))
		if err != nil {
			return nil, err
		}
		defer text.Free()
		clip := text.ClipRect
		// Dont destroy the texture. Call Destroy on the SDL_Widgets object to destroy ALL cached textures
		txt, err := renderer.CreateTextureFromSurface(text)
		if err != nil {
			return nil, err
		}
		// Crteate a new cache data
		cacheDataEntry = &SDL_TextureCacheEntry{texture: txt, clipRect: clip}
		// If there is a texture cache then add the cache entry ti it
		if tc != nil {
			tc.textureMap[cacheKey] = cacheDataEntry
		}
	}
	return cacheDataEntry, nil
}

func widgetColourDim(in *sdl.Color, doNothing bool, divBy uint8) *sdl.Color {
	r := in.R
	g := in.G
	b := in.B
	if !doNothing {
		if in.R > 5 {
			r = in.R / divBy
		}
		if in.G > 5 {
			g = in.G / divBy
		}
		if in.B > 5 {
			b = in.B / divBy
		}
	}
	return &sdl.Color{R: r, G: g, B: b, A: in.A}
}

func widgetColourBright(in *sdl.Color, doNothing bool) *sdl.Color {
	r := in.R
	g := in.G
	b := in.B
	if !doNothing {
		if in.R < 127 {
			r = in.R * 2
		}
		if in.G < 127 {
			g = in.G * 2
		}
		if in.B < 127 {
			b = in.B * 2
		}
	}
	return &sdl.Color{R: r, G: g, B: b, A: in.A}
}

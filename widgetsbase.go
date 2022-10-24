package main

import (
	"fmt"
	"math"
	"path/filepath"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type SDL_Widget interface {
	Draw(*sdl.Renderer, *ttf.Font) error
	Inside(int32, int32) bool
	GetRect() *sdl.Rect
	Click(int32, int32) bool
	SetId(int32)
	GetId() int32
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
	GetId() int32
	IsEnabled() bool
	GetForeground() *sdl.Color
}

type SDL_Shape struct {
	vx []int16
	vy []int16
}

func NewSDLShape() *SDL_Shape {
	return &SDL_Shape{vx: make([]int16, 0), vy: make([]int16, 0)}
}

func (s *SDL_Shape) Add(x, y int32) {
	s.vx = append(s.vx, int16(x))
	s.vy = append(s.vy, int16(y))
}

func (s *SDL_Shape) Rect() *sdl.Rect {
	var minx int16 = math.MaxInt16
	var miny int16 = math.MaxInt16
	var maxx int16 = math.MinInt16
	var maxy int16 = math.MinInt16
	vx := s.vx
	vy := s.vy
	for i := 0; i < len(vx); i++ {
		if vx[i] < minx {
			minx = vx[i]
		}
		if vx[i] > maxx {
			maxx = vx[i]
		}
		if vy[i] < miny {
			miny = vy[i]
		}
		if vy[i] > maxy {
			maxy = vy[i]
		}
	}
	return &sdl.Rect{X: int32(minx), Y: int32(miny), W: int32(maxx - minx), H: int32(maxy - miny)}
}

type SDL_WidgetBase struct {
	x, y, w, h int32
	id         int32
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
func initBase(x, y, w, h, id int32, deBounce int, bgColour, fgColour *sdl.Color) SDL_WidgetBase {
	return SDL_WidgetBase{x: x, y: y, w: w, h: h, id: id, enabled: true, visible: true, notPressed: true, deBounce: deBounce, bg: bgColour, fg: fgColour}
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

func (b *SDL_WidgetBase) GetRect() *sdl.Rect {
	x := b.x
	if b.w < 0 {
		x = b.x + b.w
	}
	y := b.y
	if b.h < 0 {
		y = b.y + b.h
	}
	return &sdl.Rect{X: x, Y: y, W: b.w, H: b.h}
}

func (b *SDL_WidgetBase) GetSize() (int32, int32) {
	return b.w, b.h
}
func (b *SDL_WidgetBase) SetId(id int32) {
	b.id = id
}

func (b *SDL_WidgetBase) GetId() int32 {
	return b.id
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
		r := b.GetRect()
		if x < r.X {
			return false
		}
		if y < r.Y {
			return false
		}
		if x > (r.X + r.W) {
			return false
		}
		if y > (r.Y + r.H) {
			return false
		}
		return true
	}
	return false
}

/****************************************************************************************
* Container for SDL_Widgets. A list of lists
**/
type SDL_Widget_Groups struct {
	wigets []*SDL_Widgets
}

func NewWidgetGroup() *SDL_Widget_Groups {
	return &SDL_Widget_Groups{wigets: make([]*SDL_Widgets, 0)}
}

func (wl *SDL_Widget_Groups) Add(widgets *SDL_Widgets) {
	wl.wigets = append(wl.wigets, widgets)
}

func (wl *SDL_Widget_Groups) Destroy() {
	for _, w := range wl.wigets {
		w.Destroy()
	}
}

func (wl *SDL_Widget_Groups) Draw(renderer *sdl.Renderer) {
	for _, w := range wl.wigets {
		w.Draw(renderer)
	}
}

func (wl *SDL_Widget_Groups) Inside(x, y int32) SDL_Widget {
	for _, wl := range wl.wigets {
		w := wl.Inside(x, y)
		if w != nil {
			return w
		}
	}
	return nil
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

func (wl *SDL_Widgets) LoadTextures(renderer *sdl.Renderer, applicationDataPath string, fileNames map[string]string) error {
	if wl.textureCache == nil {
		wl.textureCache = NewTextureCache()
	}
	for name, fileName := range fileNames {
		var fn string
		if applicationDataPath == "" {
			fn = fileName
		} else {
			fn = filepath.Join(applicationDataPath, fileName)
		}
		texture, rect, err := loadTextureFile(renderer, fn)
		if err != nil {
			return err
		}
		wl.textureCache.textureMap[name] = &SDL_TextureCacheEntry{name: name, texture: texture, clipRect: *rect}
	}
	return nil
}
func (wl *SDL_Widgets) GetTexture(name string) (*sdl.Texture, *sdl.Rect, error) {
	if wl.textureCache == nil {
		return nil, nil, fmt.Errorf("Texture cache is not set up")
	}
	tce := wl.textureCache.textureMap[name]
	if tce == nil {
		return nil, nil, fmt.Errorf("Texture cache does not contain %s", name)
	}
	return tce.texture, &tce.clipRect, nil
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

func (wl *SDL_Widgets) SetEnable(e bool) {
	for _, w := range wl.list {
		w.SetEnabled(e)
	}
}

func (wl *SDL_Widgets) SetVisible(e bool) {
	for _, w := range wl.list {
		w.SetVisible(e)
	}
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
	name     string
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
func loadTextureFile(renderer *sdl.Renderer, fileName string) (*sdl.Texture, *sdl.Rect, error) {
	surface, err := sdl.LoadBMP(fileName)
	if err != nil {
		return nil, nil, err
	}
	defer surface.Free()
	cRect := &surface.ClipRect
	texture, err := renderer.CreateTextureFromSurface(surface)
	if err != nil {
		return nil, nil, err
	}
	return texture, &sdl.Rect{X: cRect.X, Y: cRect.Y, W: cRect.W, H: cRect.H}, nil
}

func getCachedTextWidgetEntry(renderer *sdl.Renderer, tw SDL_TextWidget, font *ttf.Font) (*SDL_TextureCacheEntry, error) {
	// cache key must include Dim variations of the textures
	var cacheKey = fmt.Sprintf("text.%d.%s%t", tw.GetId(), tw.GetText(), tw.IsEnabled())
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
		cacheDataEntry = &SDL_TextureCacheEntry{texture: txt, clipRect: clip, name: cacheKey}
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

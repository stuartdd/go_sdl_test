package main

import (
	"fmt"
	"math"
	"path/filepath"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type ALIGN_TEXT int
type ROTATE_SHAPE_90 int

const (
	ALIGN_CENTER ALIGN_TEXT = iota
	ALIGN_LEFT
	ALIGN_RIGHT
	ALIGN_FIT

	ROTATE_0 ROTATE_SHAPE_90 = iota
	ROTATE_90
	ROTATE_180
	ROTATE_270

	DEG_TO_RAD float64 = (math.Pi / 180)
)

var TEXTURE_CACHE_TEXT_PREF = "TxCaPr987"

type SDL_Widget interface {
	Draw(*sdl.Renderer, *ttf.Font) error
	Scale(float32)
	Click(int32, int32) bool
	Inside(int32, int32) bool      // Base
	GetRect() *sdl.Rect            // Base
	SetId(int32)                   // Base
	GetId() int32                  // Base
	SetVisible(bool)               // Base
	IsVisible() bool               // Base
	SetEnabled(bool)               // Base
	IsEnabled() bool               // Base
	SetPosition(int32, int32) bool // Base
	SetSize(int32, int32) bool     // Base
	GetPosition() (int32, int32)   // Base
	GetSize() (int32, int32)       // Base
	Destroy()                      // Base
	GetForeground() *sdl.Color     // Base
	GetBackground() *sdl.Color     // Base
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

type SDL_ImageWidget interface {
	SetFrame(tf int32)
	GetFrame() int32
	NextFrame() int32
	GetFrameCount() int32
}

type SDL_TextureCacheWidget interface {
	SetTextureCache(*SDL_TextureCache)
	GetTextureCache() *SDL_TextureCache
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

func (b *SDL_WidgetBase) SetPosition(x, y int32) bool {
	if b.x != x || b.y != y {
		b.x = x
		b.y = y
		return true
	}
	return false
}

func (b *SDL_WidgetBase) GetPosition() (int32, int32) {
	return b.x, b.y
}

func (b *SDL_WidgetBase) SetSize(w, h int32) bool {
	if b.w != w || b.h != h {
		b.w = w
		b.h = h
		return true
	}
	return false
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

func (b *SDL_WidgetBase) Scale(s float32) {
	b.w = int32(float32(b.w) * s)
	b.h = int32(float32(b.h) * s)
	b.x = int32(float32(b.x) * s)
	b.y = int32(float32(b.y) * s)
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
		return isInsideRect(x, y, b.GetRect())
	}
	return false
}

/****************************************************************************************
* Container for SDL_Widgets. A list of lists
**/
type SDL_WidgetGroup struct {
	wigetLists   []*SDL_WidgetList
	font         *ttf.Font
	textureCache *SDL_TextureCache
}

func NewWidgetGroup() *SDL_WidgetGroup {
	return &SDL_WidgetGroup{wigetLists: make([]*SDL_WidgetList, 0), textureCache: NewTextureCache()}
}

func (wg *SDL_WidgetGroup) Add(widgetList *SDL_WidgetList) {
	wg.GetTextureCache().Merge(widgetList.GetTextureCache())
	widgetList.SetTextureCache(wg.GetTextureCache())
	wg.wigetLists = append(wg.wigetLists, widgetList)
}

func (wg *SDL_WidgetGroup) AllWidgets() []SDL_Widget {
	l := make([]SDL_Widget, 0)
	for _, wList := range wg.wigetLists {
		l = append(l, wList.ListWidgets()...)
	}
	return l
}

func (wg *SDL_WidgetGroup) AllLists() []*SDL_WidgetList {
	l := make([]*SDL_WidgetList, 0)
	l = append(l, wg.wigetLists...)
	return l
}

func (wg *SDL_WidgetGroup) SetFont(font *ttf.Font) {
	wg.font = font
}

func (wg *SDL_WidgetGroup) GetFont() *ttf.Font {
	return wg.font
}

func (wg *SDL_WidgetGroup) SetTextureCache(textureCache *SDL_TextureCache) {
	wg.textureCache = textureCache
}

func (wg *SDL_WidgetGroup) GetTextureCache() *SDL_TextureCache {
	return wg.textureCache
}

func (wg *SDL_WidgetGroup) LoadTexturesFromFiles(renderer *sdl.Renderer, applicationDataPath string, fileNames map[string]string) error {
	return wg.textureCache.LoadTexturesFromFiles(renderer, applicationDataPath, fileNames)
}

func (wg *SDL_WidgetGroup) LoadTexturesFromText(renderer *sdl.Renderer, textMap map[string]string, font *ttf.Font, colour *sdl.Color) error {
	return wg.textureCache.LoadTexturesFromText(renderer, textMap, font, colour)
}

func (wl *SDL_WidgetGroup) GetTexture(name string) (*sdl.Texture, *sdl.Rect, error) {
	return wl.textureCache.GetTexture(name)
}

func (wg *SDL_WidgetGroup) Scale(s float32) {
	for _, w := range wg.wigetLists {
		w.Scale(s)
	}
}

func (wg *SDL_WidgetGroup) Destroy() {
	for _, w := range wg.wigetLists {
		w.Destroy()
	}
	wg.textureCache.Destroy()
}

func (wg *SDL_WidgetGroup) Draw(renderer *sdl.Renderer) {
	for _, w := range wg.wigetLists {
		w.Draw(renderer)
	}
}

func (wg *SDL_WidgetGroup) Inside(x, y int32) SDL_Widget {
	for _, wl := range wg.wigetLists {
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
type SDL_WidgetList struct {
	list         []SDL_Widget
	font         *ttf.Font
	textureCache *SDL_TextureCache
	id           int32
}

func NewSDLWidgetList(font *ttf.Font, id int32) *SDL_WidgetList {
	return &SDL_WidgetList{textureCache: nil, list: make([]SDL_Widget, 0), font: font, id: id}
}

func (wl *SDL_WidgetList) GetId() int32 {
	return wl.id
}

func (wl *SDL_WidgetList) LoadTexturesFromFiles(renderer *sdl.Renderer, applicationDataPath string, fileNames map[string]string) error {
	if wl.textureCache == nil {
		wl.textureCache = NewTextureCache()
	}
	return wl.textureCache.LoadTexturesFromFiles(renderer, applicationDataPath, fileNames)
}

func (wl *SDL_WidgetList) GetTexture(name string) (*sdl.Texture, *sdl.Rect, error) {
	if wl.textureCache == nil {
		return nil, nil, fmt.Errorf("texture cache for SDL_WidgetList.GetTexture is nil")
	}
	return wl.textureCache.GetTexture(name)
}

func (wl *SDL_WidgetList) Add(widget SDL_Widget) {
	tw, ok := widget.(SDL_TextureCacheWidget)
	if ok {
		tw.SetTextureCache(wl.textureCache)
	}
	wl.list = append(wl.list, widget)
}

func (wl *SDL_WidgetList) Inside(x, y int32) SDL_Widget {
	for _, w := range wl.list {
		if w.Inside(x, y) {
			return w
		}
	}
	return nil
}

func (wl *SDL_WidgetList) ListWidgets() []SDL_Widget {
	return wl.list
}

func (wl *SDL_WidgetList) ArrangeLR(xx, yy, padding int32) (int32, int32) {
	x := xx
	y := yy
	var w int32
	for _, wid := range wl.list {
		if wid.IsVisible() {
			wid.SetPosition(x, y)
			w, _ = wid.GetSize()
			x = x + w + padding
		}
	}
	return x, y
}

func (wl *SDL_WidgetList) ArrangeRL(xx, yy, padding int32) (int32, int32) {
	x := xx
	y := yy
	var w int32
	for _, wid := range wl.list {
		if wid.IsVisible() {
			w, _ = wid.GetSize()
			wid.SetPosition(x-w, y)
			x = (x - w) - padding
		}
	}
	return x, y
}

func (wl *SDL_WidgetList) SetEnable(e bool) {
	for _, w := range wl.list {
		w.SetEnabled(e)
	}
}

func (wl *SDL_WidgetList) SetVisible(e bool) {
	for _, w := range wl.list {
		w.SetVisible(e)
	}
}

func (wl *SDL_WidgetList) Draw(renderer *sdl.Renderer) {
	for _, w := range wl.list {
		w.Draw(renderer, wl.font)
	}
}

func (wl *SDL_WidgetList) SetFont(font *ttf.Font) {
	wl.font = font
}

func (wl *SDL_WidgetList) GetFont() *ttf.Font {
	return wl.font
}

func (wl *SDL_WidgetList) SetTextureCache(textureCache *SDL_TextureCache) {
	wl.textureCache = textureCache
}

func (wl *SDL_WidgetList) GetTextureCache() *SDL_TextureCache {
	return wl.textureCache
}

func (wl *SDL_WidgetList) Scale(s float32) {
	for _, w := range wl.list {
		w.Scale(s)
	}
}

func (wl *SDL_WidgetList) Destroy() {
	for _, w := range wl.list {
		w.Destroy()
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
	clipRect *sdl.Rect
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

func (tc *SDL_TextureCache) Merge(fromCache *SDL_TextureCache) {
	if fromCache == nil {
		return
	}
	for n, v := range fromCache.textureMap {
		tc.textureMap[n] = v
	}
}

func (tc *SDL_TextureCache) Destroy() {
	for _, v := range tc.textureMap {
		v.Destroy()
	}
}

// func (tc *SDL_TextureCache) LoadTextures(renderer *sdl.Renderer, applicationDataPath string, fileNames map[string]string) error {
// }
func (tc *SDL_TextureCache) LoadTexturesFromText(renderer *sdl.Renderer, textMap map[string]string, font *ttf.Font, colour *sdl.Color) error {
	for name, text := range textMap {
		tce, err := getTextureCacheEntryForText(renderer, text, name, font, colour)
		if err != nil {
			return err
		}
		tc.textureMap[name] = tce
	}
	return nil
}

func (tc *SDL_TextureCache) LoadTexturesFromFiles(renderer *sdl.Renderer, applicationDataPath string, fileNames map[string]string) error {
	for name, fileName := range fileNames {
		var fn string
		if applicationDataPath == "" {
			fn = fileName
		} else {
			fn = filepath.Join(applicationDataPath, fileName)
		}
		texture, rect, err := loadTextureFile(renderer, fn)
		if err != nil {
			return fmt.Errorf("file '%s':%s", fileName, err.Error())
		}
		tc.textureMap[name] = &SDL_TextureCacheEntry{name: name, texture: texture, clipRect: rect}
	}
	return nil
}

func (tc *SDL_TextureCache) GetTexture(name string) (*sdl.Texture, *sdl.Rect, error) {
	tce := tc.textureMap[name]
	if tce == nil {
		return nil, nil, fmt.Errorf("texture cache does not contain %s", name)
	}
	return tce.texture, tce.clipRect, nil
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
	texture, err := img.LoadTexture(renderer, fileName)
	if err != nil {
		return nil, nil, err
	}
	_, _, t3, t4, err := texture.Query()
	if err != nil {
		return nil, nil, err
	}
	return texture, &sdl.Rect{X: 0, Y: 0, W: t3, H: t4}, nil
}

func getTextureCacheEntryForText(renderer *sdl.Renderer, text, name string, font *ttf.Font, colour *sdl.Color) (*SDL_TextureCacheEntry, error) {
	if colour == nil {
		colour = &sdl.Color{R: 255, G: 255, B: 255, A: 255}
	}
	surface, err := font.RenderUTF8Blended(text, *colour)
	if err != nil {
		return nil, err
	}
	defer surface.Free()
	clip := surface.ClipRect
	// Dont destroy the texture. Call Destroy on the SDL_Widgets object to destroy ALL cached textures
	txt, err := renderer.CreateTextureFromSurface(surface)
	if err != nil {
		return nil, err
	}
	return &SDL_TextureCacheEntry{texture: txt, clipRect: &clip, name: name}, nil
}

func widgetShrinkRect(in *sdl.Rect, by int32) *sdl.Rect {
	if in == nil {
		return nil
	}
	return &sdl.Rect{X: in.X + by, Y: in.Y + by, W: in.W - (by * 2), H: in.H - (by * 2)}
}

func widgetColourDim(in *sdl.Color, doNothing bool, divBy float32) *sdl.Color {
	if in == nil {
		return in
	}
	if doNothing {
		return in
	}
	return &sdl.Color{R: uint8(float32(in.R) / divBy), G: uint8(float32(in.G) / divBy), B: uint8(float32(in.B) / divBy), A: in.A}
}

func isInsideRect(x, y int32, r *sdl.Rect) bool {
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

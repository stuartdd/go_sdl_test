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
type KBD_KEY_MODE int

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

type SDL_CanFocus interface {
	SetFocus(focus bool)
	HasFocus() bool
	KeyPress(int, bool, bool) bool
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
	_enabled   bool
	notPressed bool
	deBounce   int
	bg         *sdl.Color
	fg         *sdl.Color
}

/****************************************************************************************
* Common (base) functions for ALL SDL_Widget instances
**/
func initBase(x, y, w, h, id int32, deBounce int, bgColour, fgColour *sdl.Color) SDL_WidgetBase {
	return SDL_WidgetBase{x: x, y: y, w: w, h: h, id: id, _enabled: true, visible: true, notPressed: true, deBounce: deBounce, bg: bgColour, fg: fgColour}
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
	b._enabled = e
}

func (b *SDL_WidgetBase) IsEnabled() bool {
	return b._enabled && b.notPressed && b.visible
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

func (wg *SDL_WidgetGroup) SetFocus(id int32, focus bool) {
	for _, wList := range wg.wigetLists {
		wList.SetFocus(id, focus)
	}
}

func (wg *SDL_WidgetGroup) GetFocused() SDL_CanFocus {
	for _, wList := range wg.wigetLists {
		f := wList.GetFocused()
		if f != nil {
			return f
		}
	}
	return nil
}

func (wg *SDL_WidgetGroup) AllLists() []*SDL_WidgetList {
	l := make([]*SDL_WidgetList, 0)
	l = append(l, wg.wigetLists...)
	return l
}

func (wg *SDL_WidgetGroup) SetFont(font *ttf.Font) {
	wg.font = font
}

func (wg *SDL_WidgetGroup) KeyPress(c int, ctrl, down bool) bool {
	for _, wList := range wg.wigetLists {
		if wList.KeyPress(c, ctrl, down) {
			return true
		}
	}
	return false
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

func (wg *SDL_WidgetGroup) LoadTexturesFromFileMap(renderer *sdl.Renderer, applicationDataPath string, fileMap map[string]string) error {
	return wg.textureCache.LoadTexturesFromFileMap(renderer, applicationDataPath, fileMap)
}

func (wg *SDL_WidgetGroup) LoadTexturesFromStringMap(renderer *sdl.Renderer, textMap map[string]string, font *ttf.Font, colour *sdl.Color) error {
	return wg.textureCache.LoadTexturesFromStringMap(renderer, textMap, font, colour)
}

func (wl *SDL_WidgetGroup) GetTextureForName(name string) (*sdl.Texture, int32, int32, error) {
	return wl.textureCache.GetTextureForName(name)
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
	fmt.Println(wg.textureCache.String())
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

func (wl *SDL_WidgetList) LoadTexturesFromFiles(renderer *sdl.Renderer, applicationDataPath string, fileMap map[string]string) error {
	if wl.textureCache == nil {
		wl.textureCache = NewTextureCache()
	}
	return wl.textureCache.LoadTexturesFromFileMap(renderer, applicationDataPath, fileMap)
}

func (wl *SDL_WidgetList) GetTextureForName(name string) (*sdl.Texture, int32, int32, error) {
	if wl.textureCache == nil {
		return nil, 0, 0, fmt.Errorf("texture cache for SDL_WidgetList.GetTexture is nil")
	}
	return wl.textureCache.GetTextureForName(name)
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

func (wl *SDL_WidgetList) SetFocus(id int32, focus bool) {
	for _, w := range wl.list {
		f, ok := w.(SDL_CanFocus)
		if ok {
			f.SetFocus(w.GetId() == id)
		}
	}
}

func (wl *SDL_WidgetList) GetFocused() SDL_CanFocus {
	for _, w := range wl.list {
		f, ok := w.(SDL_CanFocus)
		if ok {
			if f.HasFocus() {
				return f
			}
		}
	}
	return nil
}

func (wl *SDL_WidgetList) KeyPress(c int, ctrl, down bool) bool {
	for _, w := range wl.list {
		f, ok := w.(SDL_CanFocus)
		if ok {
			if f.HasFocus() {
				if f.KeyPress(c, ctrl, down) {
					return true
				}
			}
		}
	}
	return false
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
	if wl.textureCache != nil {
		wl.textureCache.Destroy()
	}
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
* Texture cache Entry used to hold ALL textures in the SDL_TextureCache
**/
type SDL_TextureCacheEntry struct {
	Texture *sdl.Texture
	value   string
	W, H    int32
}

func (tce *SDL_TextureCacheEntry) Destroy() int {
	if tce.Texture != nil {
		tce.Texture.Destroy()
		tce.value = ""
		return 1
	}
	return 0
}

func NewTextureCacheEntryForFile(renderer *sdl.Renderer, fileName string) (*SDL_TextureCacheEntry, error) {
	texture, err := img.LoadTexture(renderer, fileName)
	if err != nil {
		return nil, err
	}
	_, _, t3, t4, err := texture.Query()
	if err != nil {
		return nil, err
	}
	return &SDL_TextureCacheEntry{Texture: texture, W: t3, H: t4, value: fileName}, nil
}

func NewTextureCacheEntryForRune(renderer *sdl.Renderer, char rune, font *ttf.Font, colour *sdl.Color) (*SDL_TextureCacheEntry, error) {
	if colour == nil {
		colour = &sdl.Color{R: 255, G: 255, B: 255, A: 255}
	}
	surface, err := font.RenderUTF8Blended(string(char), *colour)
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
	return &SDL_TextureCacheEntry{Texture: txt, value: string(char), W: clip.W, H: clip.H}, nil
}

func NewTextureCacheEntryForString(renderer *sdl.Renderer, text string, font *ttf.Font, colour *sdl.Color) (*SDL_TextureCacheEntry, error) {
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
	return &SDL_TextureCacheEntry{Texture: txt, W: clip.W, H: clip.H}, nil
}

/****************************************************************************************
* Texture cache for widgets that have textures to display.
* Textures are sdl resources and need to be Destroyed.
* SDL_WidgetList destroys all textures via the SDL_Widgets Destroy() function.
* SDL_WidgetGroup destroys all textures via SDL_WidgetsGroup Destroy() function
* USE:		widgetGroup := NewWidgetGroup()
*       	defer widgetGroup.Destroy()
**/
type SDL_TextureCache struct {
	_textureMap map[string]*SDL_TextureCacheEntry
	in, out     int
}

func NewTextureCache() *SDL_TextureCache {
	return &SDL_TextureCache{_textureMap: make(map[string]*SDL_TextureCacheEntry), in: 0, out: 0}
}

func (tc *SDL_TextureCache) Get(name string) (*SDL_TextureCacheEntry, bool) {
	tce, ok := tc._textureMap[name]
	return tce, ok
}

func (tc *SDL_TextureCache) String() string {
	return fmt.Sprintf("TextureCache in:%d out:%d", tc.in, tc.out)
}

func (tc *SDL_TextureCache) Add(name string, tceIn *SDL_TextureCacheEntry) {
	tce := tc._textureMap[name]
	if tce != nil {
		tc.out = tc.out + tce.Destroy()
	}
	tc._textureMap[name] = tceIn
	tc.in++
}

func (tc *SDL_TextureCache) Remove(name string, tceIn *SDL_TextureCacheEntry) {
	tce := tc._textureMap[name]
	if tce != nil {
		tc.out = tc.out + tce.Destroy()
	}
	tc._textureMap[name] = nil
}

func (tc *SDL_TextureCache) Merge(fromCache *SDL_TextureCache) {
	if fromCache == nil {
		return
	}
	for n, v := range fromCache._textureMap {
		tc.Add(n, v)
	}
}

func (tc *SDL_TextureCache) Destroy() {
	for n, tce := range tc._textureMap {
		if tce != nil {
			tc.out = tc.out + tce.Destroy()
		}
		tc._textureMap[n] = nil
	}
}

func (tc *SDL_TextureCache) LoadTexturesFromStringMap(renderer *sdl.Renderer, textMap map[string]string, font *ttf.Font, colour *sdl.Color) error {
	for name, text := range textMap {
		tce, err := NewTextureCacheEntryForString(renderer, text, font, colour)
		if err != nil {
			return err
		}
		tc.Add(name, tce)
	}
	return nil
}

func (tc *SDL_TextureCache) LoadTexturesFromFileMap(renderer *sdl.Renderer, applicationDataPath string, fileNames map[string]string) error {
	for name, fileName := range fileNames {
		var fn string
		if applicationDataPath == "" {
			fn = fileName
		} else {
			fn = filepath.Join(applicationDataPath, fileName)
		}
		tce, err := NewTextureCacheEntryForFile(renderer, fn)
		if err != nil {
			return fmt.Errorf("file '%s':%s", fileName, err.Error())
		}
		tc.Add(name, tce)
	}
	return nil
}

func (tc *SDL_TextureCache) GetTextureForName(name string) (*sdl.Texture, int32, int32, error) {
	tce := tc._textureMap[name]
	if tce == nil {
		return nil, 0, 0, fmt.Errorf("texture cache does not contain %s", name)
	}
	return tce.Texture, tce.W, tce.H, nil
}

/****************************************************************************************
* Utilities
* getCachedTextWidgetEntry Returns cached texture data
*
* widgetColourDim takes a colour and returns a dimmer by same colour. Used for disabled widget text
* widgetColourBright takes a colour and returns a brighter by same colour. Used for Widget Borders
*
**/

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

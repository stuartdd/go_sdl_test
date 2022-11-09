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
type TEXT_CHANGE_TYPE int

const (
	ALIGN_CENTER ALIGN_TEXT = iota
	ALIGN_LEFT
	ALIGN_RIGHT
	ALIGN_FIT

	ROTATE_0 ROTATE_SHAPE_90 = iota
	ROTATE_90
	ROTATE_180
	ROTATE_270

	TEXT_CHANGE_INSERT TEXT_CHANGE_TYPE = iota
	TEXT_CHANGE_DELETE
	TEXT_CHANGE_BS
	TEXT_CHANGE_FENISH
	TEXT_CHANGE_NONE

	DEG_TO_RAD float64 = (math.Pi / 180)
)

var TEXTURE_CACHE_TEXT_PREF = "TxCaPr987"

type SDL_Widget interface {
	Draw(*sdl.Renderer, *ttf.Font) error
	Scale(float32)
	Click(*SDL_MouseData) bool
	Inside(int32, int32) bool      // Base
	GetRect() *sdl.Rect            // Base
	SetWidgetId(int32)             // Base
	GetWidgetId() int32            // Base
	SetVisible(bool)               // Base
	IsVisible() bool               // Base
	SetEnabled(bool)               // Base
	IsEnabled() bool               // Base
	SetPosition(int32, int32) bool // Base
	GetPosition() (int32, int32)   // Base
	SetSize(int32, int32) bool     // Base
	GetSize() (int32, int32)       // Base
	GetForeground() *sdl.Color     // Base
	GetBackground() *sdl.Color     // Base
	Destroy()                      // Base
}

type SDL_CanFocus interface {
	SetFocus(focus bool)
	HasFocus() bool
	KeyPress(int, bool, bool) bool
	ClearSelection()
}

type SDL_TextWidget interface {
	SetTextureCache(*SDL_TextureCache)
	GetTextureCache() *SDL_TextureCache
	SetText(text string)
	GetText() string
	GetWidgetId() int32
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

type SDL_MouseData struct {
	x, y, draggingToX, draggingToY int32
	button                         uint8
	down                           bool
	dragged                        bool
	dragging                       bool
	widgetId                       int32
}

type SDL_WidgetBase struct {
	x, y, w, h int32
	widgetId   int32
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
func initBase(x, y, w, h, widgetId int32, deBounce int, bgColour, fgColour *sdl.Color) SDL_WidgetBase {
	return SDL_WidgetBase{x: x, y: y, w: w, h: h, widgetId: widgetId, _enabled: true, visible: true, notPressed: true, deBounce: deBounce, bg: bgColour, fg: fgColour}
}

func (b *SDL_WidgetBase) GetWidgetId() int32 {
	return b.widgetId
}

func (b *SDL_WidgetBase) SetWidgetId(widgetId int32) {
	b.widgetId = widgetId
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
	changed := false
	if w > 0 && b.w != w {
		b.w = w
		changed = true
	}
	if h > 0 && b.h != h {
		b.h = h
		changed = true
	}
	return changed
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

func (b *SDL_WidgetBase) Scale(s float32) {
	b.w = int32(float32(b.w) * s)
	b.h = int32(float32(b.h) * s)
	b.x = int32(float32(b.x) * s)
	b.y = int32(float32(b.y) * s)
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
	wigetLists   []*SDL_WidgetSubGroup
	font         *ttf.Font
	textureCache *SDL_TextureCache
}

func NewWidgetGroup(font *ttf.Font) *SDL_WidgetGroup {
	return &SDL_WidgetGroup{wigetLists: make([]*SDL_WidgetSubGroup, 0), textureCache: NewTextureCache(), font: font}
}

func (wg *SDL_WidgetGroup) NewWidgetSubGroup(font *ttf.Font, id int32) *SDL_WidgetSubGroup {
	if font == nil {
		font = wg.font
	}
	sg := &SDL_WidgetSubGroup{textureCache: wg.textureCache, list: make([]*SDL_Widget, 0), font: font, id: id}
	wg.wigetLists = append(wg.wigetLists, sg)
	return sg
}

func (wg *SDL_WidgetGroup) AllWidgets() []*SDL_Widget {
	l := make([]*SDL_Widget, 0)
	for _, wList := range wg.wigetLists {
		l = append(l, wList.ListWidgets()...)
	}
	return l
}

func (wg *SDL_WidgetGroup) SetFocus(id int32) {
	for _, wList := range wg.wigetLists {
		wList.SetFocus(id)
	}
}

func (wg *SDL_WidgetGroup) ClearFocus() {
	for _, wList := range wg.wigetLists {
		wList.ClearFocus()
	}
}

func (wg *SDL_WidgetGroup) ClearSelection() {
	for _, wList := range wg.wigetLists {
		wList.ClearSelection()
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

func (wl *SDL_WidgetGroup) GetWidgetSubGroup(id int32) *SDL_WidgetSubGroup {
	for _, w := range wl.wigetLists {
		if (*w).GetId() == id {
			return w
		}
	}
	return nil
}

func (wl *SDL_WidgetGroup) GetWidget(id int) *SDL_Widget {
	for _, w := range wl.wigetLists {
		wi := w.GetWidget(id)
		if wi != nil {
			return wi
		}
	}
	return nil
}

func (wg *SDL_WidgetGroup) AllSubGroups() []*SDL_WidgetSubGroup {
	l := make([]*SDL_WidgetSubGroup, 0)
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
type SDL_WidgetSubGroup struct {
	id           int32
	list         []*SDL_Widget
	font         *ttf.Font
	textureCache *SDL_TextureCache
}

func (wl *SDL_WidgetSubGroup) GetId() int32 {
	return wl.id
}

func (wl *SDL_WidgetSubGroup) LoadTexturesFromFiles(renderer *sdl.Renderer, applicationDataPath string, fileMap map[string]string) error {
	if wl.textureCache == nil {
		wl.textureCache = NewTextureCache()
	}
	return wl.textureCache.LoadTexturesFromFileMap(renderer, applicationDataPath, fileMap)
}

func (wl *SDL_WidgetSubGroup) GetTextureForName(name string) (*sdl.Texture, int32, int32, error) {
	if wl.textureCache == nil {
		return nil, 0, 0, fmt.Errorf("texture cache for SDL_WidgetList.GetTexture is nil")
	}
	return wl.textureCache.GetTextureForName(name)
}

func (wl *SDL_WidgetSubGroup) Add(widget SDL_Widget) {
	tw, ok := widget.(SDL_TextureCacheWidget)
	if ok {
		tw.SetTextureCache(wl.textureCache)
	}
	wl.list = append(wl.list, &widget)
}

func (wl *SDL_WidgetSubGroup) Inside(x, y int32) SDL_Widget {
	for _, w := range wl.list {
		if (*w).Inside(x, y) {
			return (*w)
		}
	}
	return nil
}

func (wl *SDL_WidgetSubGroup) ListWidgets() []*SDL_Widget {
	return wl.list
}

func (wl *SDL_WidgetSubGroup) GetWidget(id int) *SDL_Widget {
	for _, w := range wl.list {
		if int((*w).GetWidgetId()) == id {
			return w
		}
	}
	return nil
}

func (wl *SDL_WidgetSubGroup) SetFocus(id int32) {
	for _, w := range wl.list {
		f, ok := (*w).(SDL_CanFocus)
		if ok {
			f.SetFocus((*w).GetWidgetId() == id)
		}
	}
}

func (wl *SDL_WidgetSubGroup) ClearFocus() {
	for _, w := range wl.list {
		f, ok := (*w).(SDL_CanFocus)
		if ok {
			f.SetFocus(false)
		}
	}
}

func (wl *SDL_WidgetSubGroup) ClearSelection() {
	for _, w := range wl.list {
		f, ok := (*w).(SDL_CanFocus)
		if ok {
			f.ClearSelection()
		}
	}
}
func (wl *SDL_WidgetSubGroup) GetFocused() SDL_CanFocus {
	for _, w := range wl.list {
		f, ok := (*w).(SDL_CanFocus)
		if ok {
			if f.HasFocus() {
				return f
			}
		}
	}
	return nil
}

func (wl *SDL_WidgetSubGroup) KeyPress(c int, ctrl, down bool) bool {
	for _, w := range wl.list {
		f, ok := (*w).(SDL_CanFocus)
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

func (wl *SDL_WidgetSubGroup) ArrangeLR(xx, yy, padding int32) (int32, int32) {
	x := xx
	y := yy
	var width int32
	var w *SDL_Widget
	for _, w = range wl.list {
		if (*w).IsVisible() {
			(*w).SetPosition(x, y)
			width, _ = (*w).GetSize()
			x = x + width + padding
		}
	}
	return x, y
}

func (wl *SDL_WidgetSubGroup) ArrangeRL(xx, yy, padding int32) (int32, int32) {
	x := xx
	y := yy
	var width int32
	for _, w := range wl.list {
		if (*w).IsVisible() {
			width, _ = (*w).GetSize()
			(*w).SetPosition(x-width, y)
			x = (x - width) - padding
		}
	}
	return x, y
}

func (wl *SDL_WidgetSubGroup) SetEnable(e bool) {
	for _, w := range wl.list {
		(*w).SetEnabled(e)
	}
}

func (wl *SDL_WidgetSubGroup) SetVisible(e bool) {
	for _, w := range wl.list {
		(*w).SetVisible(e)
	}
}

func (wl *SDL_WidgetSubGroup) Draw(renderer *sdl.Renderer) {
	for _, w := range wl.list {
		(*w).Draw(renderer, wl.font)
	}
}

func (wl *SDL_WidgetSubGroup) SetFont(font *ttf.Font) {
	wl.font = font
}

func (wl *SDL_WidgetSubGroup) GetFont() *ttf.Font {
	return wl.font
}

func (wl *SDL_WidgetSubGroup) Scale(s float32) {
	for _, w := range wl.list {
		(*w).Scale(s)
	}
}

func (wl *SDL_WidgetSubGroup) Destroy() {
	for _, w := range wl.list {
		(*w).Destroy()
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

func (tc *SDL_TextureCache) String() string {
	return fmt.Sprintf("TextureCache in:%d out:%d", tc.in, tc.out)
}

func (tc *SDL_TextureCache) Peek(name string) bool {
	_, ok := tc._textureMap[name]
	return ok
}

func (tc *SDL_TextureCache) Get(name string) (*SDL_TextureCacheEntry, bool) {
	tce, ok := tc._textureMap[name]
	return tce, ok
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

func (md *SDL_MouseData) String() string {
	return fmt.Sprintf("x:%d y:%d Dx:%d Dy:%d ID:%d Btn:%d Down:%t Dragged:%t Dragging:%t", md.GetX(), md.GetY(), md.draggingToX, md.draggingToY, md.widgetId, md.button, md.IsDown(), md.IsDragged(), md.IsDragging())
}

func (md *SDL_MouseData) IsDown() bool {
	return md.down
}

func (md *SDL_MouseData) IsDragged() bool {
	return md.dragged
}

func (md *SDL_MouseData) IsDragging() bool {
	return md.dragging
}

func (md *SDL_MouseData) SetDragged(d bool) {
	md.dragging = false
	md.dragged = d
}

func (md *SDL_MouseData) SetDragging(d bool) {
	if d {
		md.dragging = true
	} else {
		md.draggingToX = 0
		md.draggingToY = 0
		md.dragging = false
		md.widgetId = 0
	}
	md.dragged = false
}

func (md *SDL_MouseData) GetWidgetId() int32 {
	return md.widgetId
}

func (md *SDL_MouseData) SetWidgetId(id int32) {
	if md.widgetId != id {
		md.draggingToX = 0
		md.draggingToY = 0
		md.dragging = false
		md.dragged = false
	}
	md.widgetId = id
}

func (md *SDL_MouseData) GetY() int32 {
	return md.y
}

func (md *SDL_MouseData) GetX() int32 {
	return md.x
}

func (md *SDL_MouseData) GetButtons() uint8 {
	return md.button
}

func (md *SDL_MouseData) SetButtons(b uint8) {
	md.button = b
}

func (md *SDL_MouseData) SetXY(x, y int32) {
	if md.dragging {
		md.draggingToX = x
		md.draggingToY = y
	} else {
		md.x = x
		md.y = y
	}
	md.down = true
	md.dragged = false
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

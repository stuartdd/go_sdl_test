package main

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

/****************************************************************************************
* SDL_Label code
* Implements SDL_Widget cos it is one!
* Implements SDL_TextWidget because it has text and uses the texture cache
**/
type SDL_Entry struct {
	SDL_WidgetBase
	text     string
	cursor   int
	leadin   int
	leadout  int
	hasfocus bool
	list     []*SDL_EntryChar
}

type SDL_EntryChar struct {
	texture *sdl.Texture
	text    string
	w, h    int32
	offset  int32
}

var _ SDL_Widget = (*SDL_Entry)(nil)   // Ensure SDL_Button 'is a' SDL_Widget
var _ SDL_CanFocus = (*SDL_Entry)(nil) // Ensure SDL_Button 'is a' SDL_Widget

func newEntryChar(renderer *sdl.Renderer, font *ttf.Font, colour *sdl.Color, text string) (*SDL_EntryChar, error) {
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
	return &SDL_EntryChar{texture: txt, text: text, w: clip.W, h: clip.H, offset: 0}, nil
}

func NewSDLEntry(x, y, w, h, id int32, text string, bgColour, fgColour *sdl.Color) *SDL_Entry {
	but := &SDL_Entry{text: text, list: nil, cursor: len(text), leadin: 0, leadout: 0, hasfocus: true}
	but.SDL_WidgetBase = initBase(x, y, w, h, id, 0, bgColour, fgColour)
	return but
}

func (b *SDL_Entry) SetForeground(c *sdl.Color) {
	if b.fg != c {
		b.list = nil
		b.fg = c
	}
}

func (b *SDL_Entry) SetText(text string) {
	if b.text != text {
		b.list = nil
		b.text = text
	}
}

func (b *SDL_Entry) SetFocus(focus bool) {
	if b.IsEnabled() {
		b.hasfocus = focus
	} else {
		b.hasfocus = false
	}
}

func (b *SDL_Entry) HasFocus() bool {
	if b.IsEnabled() {
		return b.hasfocus
	} else {
		return false
	}
}

func (b *SDL_Entry) KeyPress(c byte, ws bool) bool {
	if b.IsEnabled() {
		if ws {
			fmt.Printf("Key: %d\n", c)
		} else {
			fmt.Printf("Key: '%c'\n", c)
		}
		return true
	} else {
		return false
	}
}

func (b *SDL_Entry) MoveCursor(i int) {
	if b.HasFocus() {
		if i < 0 {
			if (b.cursor + i) >= 0 {
				b.cursor = b.cursor + i
			}
		}
		if b.cursor < b.leadin {
			b.leadin = b.cursor
		}
		if i > 0 {
			if (b.cursor + i) <= len(b.text) {
				b.cursor = b.cursor + i
			}
			if b.cursor > b.leadout && b.leadout < len(b.text) {
				b.leadin = b.leadin + 1
			}
		}
	}
}

func (b *SDL_Entry) SetEnabled(e bool) {
	if b.IsEnabled() != e {
		b.list = nil
		b.SDL_WidgetBase.SetEnabled(e)
	}
}

func (b *SDL_Entry) GetText() string {
	return b.text
}

func (b *SDL_Entry) Click(x, y int32) bool {
	if b.IsEnabled() {
		if b.list != nil {
			for i := b.leadin; i < len(b.text); i++ {
				ec := b.list[i]
				if x < (ec.offset + ec.w) {
					b.cursor = i
					return true
				}
			}
			b.cursor = b.leadout
		}
	}
	return false
}

func (b *SDL_Entry) buildImage(renderer *sdl.Renderer, font *ttf.Font, colour *sdl.Color, text string) ([]*SDL_EntryChar, error) {
	fmt.Println("buildImage")
	list := make([]*SDL_EntryChar, 0)
	for _, c := range text {
		ec, err := newEntryChar(renderer, font, colour, string(c))
		if err != nil {
			return nil, err
		}
		list = append(list, ec)
	}
	return list, nil
}

func (b *SDL_Entry) Draw(renderer *sdl.Renderer, font *ttf.Font) error {
	if b.IsVisible() {
		var err error
		if b.list == nil {
			b.list, err = b.buildImage(renderer, font, b.fg, b.text)
			if err != nil {
				renderer.SetDrawColor(255, 0, 0, 255)
				renderer.DrawRect(&sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
				return nil
			}
		}
		inset := float32(b.h) / 4
		th := float32(b.h) - inset
		ty := (float32(b.h) - th) / 2
		tx := b.x + 10
		if b.bg != nil {
			renderer.SetDrawColor(b.bg.R, b.bg.G, b.bg.B, b.bg.A)
			renderer.FillRect(&sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
		}
		cursorNotDrawn := true
		for i := b.leadin; i < len(b.text); i++ {
			ec := b.list[i]
			renderer.Copy(ec.texture, nil, &sdl.Rect{X: tx, Y: b.y + int32(ty), W: ec.w, H: ec.h})
			if b.IsEnabled() {
				if i == b.cursor {
					renderer.SetDrawColor(255, 255, 255, 255)
					renderer.FillRect(&sdl.Rect{X: tx, Y: b.y, W: 2, H: b.h})
					cursorNotDrawn = false
				}
				ec.offset = tx
				b.leadout = i + 1
			}
			tx = tx + ec.w
			if tx+ec.w >= b.x+b.w {
				break
			}
		}
		if cursorNotDrawn && b.IsEnabled() {
			renderer.SetDrawColor(255, 255, 255, 255)
			renderer.FillRect(&sdl.Rect{X: tx, Y: b.y, W: 2, H: b.h})
		}
		if b.fg != nil {
			borderColour := widgetColourDim(b.fg, b.IsEnabled(), 2)
			renderer.SetDrawColor(borderColour.R, borderColour.G, borderColour.B, borderColour.A)
			renderer.DrawRect(&sdl.Rect{X: b.x + 1, Y: b.y + 1, W: b.w - 1, H: b.h - 1})
		}
	}
	return nil
}
func (b *SDL_Entry) Destroy() {
	// Image cache takes care of all images!
}

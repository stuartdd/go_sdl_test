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
	text        string
	history     []string
	cursor      int
	cursorTimer int
	hasfocus    bool
	ctrlKeyDown bool
	list        []*SDL_EntryChar
	leadin      int
	leadout     int
}

type SDL_EntryChar struct {
	texture *sdl.Texture
	char    rune
	w, h    int32
	offset  int32
}

var _ SDL_Widget = (*SDL_Entry)(nil)   // Ensure SDL_Button 'is a' SDL_Widget
var _ SDL_CanFocus = (*SDL_Entry)(nil) // Ensure SDL_Button 'is a' SDL_Widget

func newEntryChar(renderer *sdl.Renderer, font *ttf.Font, colour *sdl.Color, char rune) (*SDL_EntryChar, error) {
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
	return &SDL_EntryChar{texture: txt, char: char, w: clip.W, h: clip.H, offset: 0}, nil
}

func (ec *SDL_EntryChar) String() string {
	return fmt.Sprintf("'%c' w:%d h:%d offset:%d", ec.char, ec.w, ec.h, ec.offset)
}

func NewSDLEntry(x, y, w, h, id int32, text string, bgColour, fgColour *sdl.Color) *SDL_Entry {
	but := &SDL_Entry{text: text, list: nil, cursor: 0, cursorTimer: 0, leadin: 0, leadout: 0, hasfocus: false, ctrlKeyDown: false}
	but.SDL_WidgetBase = initBase(x, y, w, h, id, 0, bgColour, fgColour)
	return but
}

func (b *SDL_Entry) SetForeground(c *sdl.Color) {
	if b.fg != c {
		b.fg = c
		b.invalidate()
	}
}

func (b *SDL_Entry) pushHistory() {
	if len(b.history) > 0 {
		if (b.history)[len(b.history)-1] == b.text {
			return
		}
	}
	b.history = append(b.history, b.text)
}

func (b *SDL_Entry) popHistory() bool {
	if len(b.history) > 0 {
		b.text = (b.history)[len(b.history)-1]
		b.history = (b.history)[0 : len(b.history)-1]
		b.invalidate()
		return true
	}
	return false
}

func (b *SDL_Entry) SetText(text string) {
	if b.text != text {
		b.text = text
		b.invalidate()
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

func (b *SDL_Entry) KeyPress(c int, ctrl bool, down bool) bool {
	if b.IsEnabled() && b.HasFocus() {
		if ctrl {
			// if ctrl key then just remember its state (up or down)
			if c == sdl.K_LCTRL || c == sdl.K_RCTRL {
				b.ctrlKeyDown = down
				return true
			}
			// if the control key is down then it is a control sequence like CTRL-Z
			if b.ctrlKeyDown {
				b.popHistory()
				b.SetCursor(b.cursor) // Ensure cursor is in range
				return true
			}
			// If it is NOT a ctrl key or a control sequence then we only react on a DOWN
			if down {
				if c < 32 || c == 127 {
					switch c {
					case sdl.K_DELETE:
						if b.cursor < len(b.text) {
							b.pushHistory()
							b.text = fmt.Sprintf("%s%s", b.text[0:b.cursor], b.text[b.cursor+1:])
							b.invalidate()
						}
					case sdl.K_BACKSPACE:
						if b.cursor > 0 {
							b.pushHistory()
							if b.cursor < len(b.text) {
								b.text = fmt.Sprintf("%s%s", b.text[0:b.cursor-1], b.text[b.cursor:])
							} else {
								b.text = b.text[0 : len(b.text)-1]
							}
							b.MoveCursor(-1)
							b.invalidate()
						}
					case sdl.K_RETURN:
						fmt.Println("CR")
					default:
						fmt.Printf("??:%d", c)
						return false
					}
				} else {
					switch c | 0x40000000 {
					case sdl.K_RIGHT:
						b.MoveCursor(1)
					case sdl.K_UP:
						b.SetCursor(99)
					case sdl.K_DOWN:
						b.SetCursor(0)
					case sdl.K_LEFT:
						b.MoveCursor(-1)
					default:
						return false
					}
				}
			} else {
				// If it is NOT a ctrl key or a control sequence then we ignore an UP
				return false
			}
		} else {
			// not a control key. insert it at the cursor
			b.pushHistory()
			b.text = fmt.Sprintf("%s%c%s", b.text[0:b.cursor], c, b.text[b.cursor:])
			b.MoveCursor(1)
			b.invalidate()
		}
		return true
	} else {
		return false
	}
}

func (b *SDL_Entry) SetCursor(i int) {
	if b.HasFocus() {
		if i < 0 {
			i = 0
		}
		if i > len(b.text) {
			i = len(b.text)
		}
		b.cursor = i
	}
}

func (b *SDL_Entry) MoveCursor(i int) {
	b.SetCursor(b.cursor + i)
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
					b.SetCursor(i)
					return true
				}
			}
			b.SetCursor(b.leadout)
		}
	}
	return false
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
		if b.bg != nil {
			renderer.SetDrawColor(b.bg.R, b.bg.G, b.bg.B, b.bg.A)
			renderer.FillRect(&sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
		}
		inset := float32(b.h) / 4
		th := float32(b.h) - inset
		ty := (float32(b.h) - th) / 2

		// *******************************************************
		tx := b.x + 10
		cc := 0
		for pos := b.leadin; pos < len(b.text); pos++ {
			ec := b.list[pos]
			ec.offset = tx
			tx = tx + ec.w
			cc++
			if tx+ec.w >= b.x+b.w {
				break
			}
		}

		if b.cursor < b.leadin {
			b.leadin = b.cursor
		}
		b.leadout = b.leadin + cc

		if b.cursor > b.leadout {
			b.leadin = b.cursor - cc
			if b.leadin < 0 {
				b.leadin = 0
			}
			b.leadout = b.leadin + cc
		}

		//*********************************************************
		tx = b.x + 10
		cursorNotVisible := true
		paintCursor := b.IsEnabled() && b.HasFocus() && (sdl.GetTicks64()%1000) > 300
		for pos := b.leadin; pos < b.leadout; pos++ {
			ec := b.list[pos]
			renderer.Copy(ec.texture, nil, &sdl.Rect{X: tx, Y: b.y + int32(ty), W: ec.w, H: ec.h})
			if paintCursor {
				if pos == b.cursor {
					renderer.SetDrawColor(255, 255, 255, 255)
					renderer.FillRect(&sdl.Rect{X: tx, Y: b.y, W: 2, H: b.h})
					cursorNotVisible = false
				}
			}
			tx = tx + ec.w
		}
		if cursorNotVisible && paintCursor {
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

func (b *SDL_Entry) buildImage(renderer *sdl.Renderer, font *ttf.Font, colour *sdl.Color, text string) ([]*SDL_EntryChar, error) {
	list := make([]*SDL_EntryChar, 0)
	for _, c := range text {
		ec, err := newEntryChar(renderer, font, colour, c)
		if err != nil {
			return nil, err
		}
		list = append(list, ec)
	}
	return list, nil
}

func (b *SDL_Entry) invalidate() {
	b.list = nil
}

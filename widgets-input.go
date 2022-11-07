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
	text         string
	textLen      int
	history      []string
	cursor       int
	cursorTimer  int
	hasfocus     bool
	ctrlKeyDown  bool
	textureCache *SDL_TextureCache
	invalid      bool
	errorState   error
	leadin       int
	leadout      int
	indent       int32
	onChange     func(string, string, TEXT_CHANGE_TYPE) (string, error)
}

var _ SDL_Widget = (*SDL_Entry)(nil)   // Ensure SDL_Button 'is a' SDL_Widget
var _ SDL_CanFocus = (*SDL_Entry)(nil) // Ensure SDL_Button 'is a' SDL_Widget

func NewSDLEntry(x, y, w, h, id int32, text string, bgColour, fgColour *sdl.Color, onChange func(string, string, TEXT_CHANGE_TYPE) (string, error)) *SDL_Entry {
	but := &SDL_Entry{text: text, textLen: len(text), textureCache: nil, cursor: 0, cursorTimer: 0, leadin: 0, leadout: 0, hasfocus: false, ctrlKeyDown: false, invalid: true, indent: 10, onChange: onChange, errorState: nil}
	but.SDL_WidgetBase = initBase(x, y, w, h, id, 0, bgColour, fgColour)
	return but
}

func (b *SDL_Entry) SetTextureCache(tc *SDL_TextureCache) {
	b.textureCache = tc
}

func (b *SDL_Entry) GetTextureCache() *SDL_TextureCache {
	return b.textureCache
}

func (b *SDL_Entry) SetForeground(c *sdl.Color) {
	if b.fg != c {
		b.fg = c
		b.invalidate()
	}
}

func (b *SDL_Entry) SetErrorState(err error) {
	b.errorState = err
}

func (b *SDL_Entry) pushHistory(val string) {
	if len(b.history) > 0 {
		if (b.history)[len(b.history)-1] == val {
			return
		}
	}
	b.history = append(b.history, val)
}

func (b *SDL_Entry) SetText(text string) {
	if b.text != text {
		b.text = text
		b.textLen = len(b.text)
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
		var err error
		oldValue := b.text
		newValue := b.text
		onChangeType := TEXT_CHANGE_NONE
		saveHistory := true
		if ctrl {
			// if ctrl key then just remember its state (up or down)
			if c == sdl.K_LCTRL || c == sdl.K_RCTRL {
				b.ctrlKeyDown = down
				return true
			}
			// if the control key is down then it is a control sequence like CTRL-Z
			if b.ctrlKeyDown {
				if c == sdl.K_z {
					if len(b.history) > 0 {
						newValue = (b.history)[len(b.history)-1]
						b.history = (b.history)[0 : len(b.history)-1]
						saveHistory = false
					}
				}
			} else {
				if down {
					if c < 32 || c == 127 {
						switch c {
						case sdl.K_DELETE:
							if b.cursor < b.textLen {
								newValue = fmt.Sprintf("%s%s", oldValue[0:b.cursor], oldValue[b.cursor+1:])
								onChangeType = TEXT_CHANGE_DELETE
							}
						case sdl.K_BACKSPACE:
							if b.cursor > 0 {
								if b.cursor < b.textLen {
									newValue = fmt.Sprintf("%s%s", oldValue[0:b.cursor-1], oldValue[b.cursor:])
								} else {
									newValue = oldValue[0 : b.textLen-1]
								}
								onChangeType = TEXT_CHANGE_BS
							}
						case sdl.K_RETURN:
							if b.onChange != nil {
								b.onChange("", b.text, TEXT_CHANGE_FENISH)
							}
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
					// If it is NOT down then we ignore an
					return false
				}
			}
			// If it is NOT a ctrl key or a control sequence then we only react on a DOWN

		} else {
			// not a control key. insert it at the cursor
			newValue = fmt.Sprintf("%s%c%s", oldValue[0:b.cursor], c, oldValue[b.cursor:])
			onChangeType = TEXT_CHANGE_INSERT
		}
		if oldValue != newValue && b.onChange != nil {
			newValue, err = b.onChange(oldValue, newValue, onChangeType)
			b.SetErrorState(err)
		}
		if newValue != oldValue {
			if saveHistory {
				b.pushHistory(oldValue)
			}
			b.text = newValue
			b.textLen = len(b.text)
			switch onChangeType {
			case TEXT_CHANGE_INSERT:
				b.MoveCursor(1)
			case TEXT_CHANGE_BS:
				b.MoveCursor(-1)
			}
			b.invalidate()
			return true
		}
	}
	return false
}

func (b *SDL_Entry) SetCursor(i int) {
	if b.HasFocus() {
		if i < 0 {
			i = 0
		}
		if i > b.textLen {
			i = b.textLen
		}
		b.cursor = i
	}
}

func (b *SDL_Entry) MoveCursor(i int) {
	b.SetCursor(b.cursor + i)
}

func (b *SDL_Entry) SetEnabled(e bool) {
	if b.IsEnabled() != e {
		b.invalidate()
		b.SDL_WidgetBase.SetEnabled(e)
	}
}

func (b *SDL_Entry) GetText() string {
	return b.text
}

func (b *SDL_Entry) Click(x, y int32) bool {
	if b.IsEnabled() {
		list := b.getTextureListFromCache()
		cur := b.x + b.indent
		for pos := b.leadin; pos < b.leadout; pos++ {
			ec := list[pos]
			if cur > x {
				b.SetCursor(pos)
				return true
			}
			cur = cur + ec.W
		}
		b.SetCursor(b.leadout)
	}
	return false
}

func (b *SDL_Entry) Draw(renderer *sdl.Renderer, font *ttf.Font) error {
	if b.IsVisible() {
		var err error
		var ec *SDL_TextureCacheEntry
		if b.invalid {
			err = b.updateTextureCache(renderer, font, b.fg, b.text)
			if err != nil {
				renderer.SetDrawColor(255, 0, 0, 255)
				renderer.DrawRect(&sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
				return nil
			}
		}

		// *******************************************************
		// Find the number of chars thet can be displayed 'cc'
		tx := b.x + b.indent
		cc := 0
		if b.leadin < 0 { // Ensure leadin is not negative
			b.leadin = 0
		}
		b.leadout = b.textLen
		list := b.getTextureListFromCache() // Get the textures and their widths
		// work out how many chars will fit in the rectangle
		for pos := b.leadin; pos < b.textLen; pos++ {
			ec = list[pos]
			tx = tx + ec.W
			if tx >= b.x+b.w {
				break
			}
			cc++
		}

		if b.leadin > 0 && cc == 0 {
			b.leadin--
			cc++
		}

		if b.leadin > 0 && b.leadin == b.cursor {
			b.leadin--
		}
		b.leadout = b.leadin + cc
		if b.cursor > b.leadout {
			b.leadin = b.cursor - cc
			b.leadout = b.leadin + cc
		}

		//*********************************************************
		if b.bg != nil {
			if b.errorState != nil {
				renderer.SetDrawColor(100, 0, 0, b.bg.A)
			} else {
				renderer.SetDrawColor(b.bg.R, b.bg.G, b.bg.B, b.bg.A)
			}
			renderer.FillRect(&sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
		}
		tx = b.x + int32(b.indent)
		th := float32(b.h) - float32(b.h)/4
		ty := (float32(b.h) - th) / 2

		cursorNotVisible := true
		paintCursor := b.IsEnabled() && b.HasFocus() && (sdl.GetTicks64()%1000) > 300
		for pos := b.leadin; pos < b.leadout; pos++ {
			ec := list[pos]
			renderer.Copy(ec.Texture, nil, &sdl.Rect{X: tx, Y: b.y + int32(ty), W: ec.W, H: ec.H})
			if paintCursor {
				if pos == b.cursor {
					renderer.SetDrawColor(255, 255, 255, 255)
					renderer.FillRect(&sdl.Rect{X: tx, Y: b.y, W: 2, H: b.h})
					cursorNotVisible = false
				}
			}
			tx = tx + ec.W
		}
		if cursorNotVisible && paintCursor {
			renderer.SetDrawColor(255, 255, 255, 255)
			renderer.FillRect(&sdl.Rect{X: tx, Y: b.y, W: 2, H: b.h})
		}
		if b.hasfocus {
			renderer.SetDrawColor(255, 0, 0, 255)
			renderer.DrawRect(&sdl.Rect{X: b.x + 1, Y: b.y + 1, W: b.w - 2, H: b.h - 2})
		} else {
			if b.fg != nil {
				borderColour := widgetColourDim(b.fg, b.IsEnabled(), 2)
				renderer.SetDrawColor(borderColour.R, borderColour.G, borderColour.B, borderColour.A)
				renderer.DrawRect(&sdl.Rect{X: b.x + 1, Y: b.y + 1, W: b.w - 2, H: b.h - 2})
			}
		}
	}
	return nil
}
func (b *SDL_Entry) Destroy() {
	// Image cache takes care of all images!
}

func (b *SDL_Entry) updateTextureCache(renderer *sdl.Renderer, font *ttf.Font, colour *sdl.Color, text string) error {
	for _, c := range text {
		key := string(c) + "-cHaR"
		ok := b.textureCache.Peek(key)
		if !ok {
			ec, err := NewTextureCacheEntryForRune(renderer, c, font, colour)
			if err != nil {
				return err
			}
			b.textureCache.Add(key, ec)
		}
	}
	return nil
}

func (b *SDL_Entry) getTextureListFromCache() []*SDL_TextureCacheEntry {
	list := make([]*SDL_TextureCacheEntry, len(b.text))
	for i, c := range b.text {
		ec, _ := b.textureCache.Get(string(c) + "-cHaR")
		list[i] = ec
	}
	return list
}

func (b *SDL_Entry) invalidate() {
	go func() {
		sdl.Delay(100)
		b.invalid = true
	}()
}

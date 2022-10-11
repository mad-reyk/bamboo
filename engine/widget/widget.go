package widget

import (
	"github.com/3elDU/bamboo/engine"
	"github.com/3elDU/bamboo/engine/texture"
	"github.com/veandco/go-sdl2/sdl"
)

type Anchor int

const (
	TopLeft Anchor = iota
	Top
	TopRight

	Left
	Center
	Right

	BottomLeft
	Bottom
	BottomRight
)

type Widget interface {
	Anchor() Anchor
	Render() *sdl.Texture
}

func RenderSingle(engine *engine.Engine, widget Widget) {
	tex := widget.Render()
	// HACK: A temporary workaround. WIll break a lot of widgets in future
	defer tex.Destroy()

	width, height := texture.Dimensions(tex)
	windowWidth, windowHeight := engine.Win.GetSize()

	var rect sdl.Rect
	switch widget.Anchor() {
	case TopLeft:
		rect = sdl.Rect{
			X: 0, Y: 0,
			W: width, H: height,
		}
	case Top:
		rect = sdl.Rect{
			X: windowWidth/2 - width/2, Y: 0,
			W: width, H: height,
		}
	case TopRight:
		rect = sdl.Rect{
			X: windowWidth - width, Y: 0,
			W: width, H: height,
		}

	case Left:
		rect = sdl.Rect{
			X: 0, Y: windowHeight/2 - height/2,
			W: width, H: height,
		}
	case Center:
		rect = sdl.Rect{
			X: windowWidth/2 - width/2, Y: windowHeight/2 - height/2,
			W: width, H: height,
		}
	case Right:
		rect = sdl.Rect{
			X: windowWidth - width, Y: windowHeight/2 - height/2,
			W: width, H: height,
		}

	case BottomLeft:
		rect = sdl.Rect{
			X: 0, Y: windowHeight - height,
			W: width, H: height,
		}
	case Bottom:
		rect = sdl.Rect{
			X: windowWidth/2 - width/2, Y: windowHeight - height,
			W: width, H: height,
		}
	case BottomRight:
		rect = sdl.Rect{
			X: windowWidth - width, Y: windowHeight - height,
			W: width, H: height,
		}
	}

	engine.Ren.Copy(tex, nil, &rect)
}

func RenderMultiple(engine *engine.Engine, widgets []Widget) {
	for _, widget := range widgets {
		RenderSingle(engine, widget)
	}
}

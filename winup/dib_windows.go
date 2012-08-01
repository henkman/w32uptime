package main

import (
	"code.google.com/p/draw2d/draw2d"
	"github.com/skelterjohn/go.wde/win"
	"code.google.com/p/freetype-go/freetype/raster"
	"image/color"
	"image/draw"	
)

func NewGraphicContext(img draw.Image) *draw2d.ImageGraphicContext {
	return draw2d.NewGraphicContextWithPainter(img, NewDIBPainter(img.(*win.DIB)))
}

func NewDIBPainter(m *win.DIB) draw2d.Painter {
	return &DIBPainter{Image: m}
}

type DIBPainter struct {
	// The image to compose onto.
	Image *win.DIB
	// The Porter-Duff composition operator.
	Op draw.Op
	// The 16-bit color to paint the spans.
	cr, cg, cb uint32
}

// Paint satisfies the Painter interface by painting ss onto a win.DIB.
func (r *DIBPainter) Paint(ss []raster.Span, done bool) {
	b := r.Image.Bounds()
	for _, s := range ss {
		if s.Y < b.Min.Y {
			continue
		}
		if s.Y >= b.Max.Y {
			return
		}
		if s.X0 < b.Min.X {
			s.X0 = b.Min.X
		}
		if s.X1 > b.Max.X {
			s.X1 = b.Max.X
		}
		if s.X0 >= s.X1 {
			continue
		}
	
		i0 :=(r.Image.Rect.Max.Y-s.Y-r.Image.Rect.Min.Y-1)*r.Image.Stride + (s.X0-r.Image.Rect.Min.X)*3
		i1 := i0 + (s.X1-s.X0) * 3
		for i := i0; i < i1; i += 3 {
			r.Image.Pix[i+0] = uint8(r.cb)
			r.Image.Pix[i+1] = uint8(r.cg)
			r.Image.Pix[i+2] = uint8(r.cr)
		}
	}
}

// SetColor sets the color to paint the spans.
func (r *DIBPainter) SetColor(c color.Color) {
	r.cr, r.cg, r.cb, _ = c.RGBA()
}

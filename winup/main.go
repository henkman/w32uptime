package main

import (
	"code.google.com/p/draw2d/draw2d"
	"github.com/papplampe/w32uptime"
	"github.com/skelterjohn/go.wde"
	_ "github.com/skelterjohn/go.wde/init"
	"image/color"
	"time"
)

const (
	Day          = time.Duration(24) * time.Hour
	BAR_HEIGHT   = 60
	BAR_WIDTH    = 24 * 20
	MIN_DIVISOR  = (24 * 60) / BAR_WIDTH
	IA_START_Y   = 40
	IA_BUFFER    = 10
	FONT_START_Y = 60 + IA_START_Y + IA_BUFFER
	FONT_START_X = 15
	BAR_START_Y  = FONT_START_Y - 45
	BAR_START_X  = FONT_START_X + 95
	WIN_WIDTH    = BAR_WIDTH + BAR_START_X + SB_SIZE + 30
	WIN_HEIGHT   = BAR_START_Y + 7*BAR_HEIGHT + 10
	SB_START_X   = BAR_WIDTH + BAR_START_X + 10
	SB_SIZE      = 30
	DATE_FORMAT  = "02 Jan 2006"
)

var (
	uptimes      []w32uptime.Uptime
	arrowButtons []ArrowButton
	curTime      time.Time
)

type ArrowButton struct {
	X, Y    int
	W, H    int
	Up      bool
	OnClick func(sb ArrowButton)
}

func (this ArrowButton) Click() {
	this.OnClick(this)
}

func (this ArrowButton) Paint(gc draw2d.GraphicContext) {
	gc.SetFillColor(color.RGBA{0x22, 0x22, 0x22, 0xFF})
	rect(gc, float64(this.X), float64(this.X+this.W), float64(this.Y), float64(this.Y+this.H))
	gc.Fill()
	gc.SetFillColor(color.RGBA{0xFF, 0x33, 0x33, 0xFF})
	arrow(gc, float64(this.X), float64(this.Y), float64(this.W), float64(this.H), this.Up)
	gc.Fill()
}

func (this ArrowButton) Contains(x, y int) bool {
	return x > this.X && x < this.X+this.W && y > this.Y && y < this.Y+this.H
}

func isUptimeInRange(s, e time.Time, up w32uptime.Uptime) bool {
	return (up.Start.After(s) && up.Start.Before(e)) ||
		(up.End.After(s) && up.End.Before(e)) ||
		(up.Start.Before(s) && up.End.After(e))
}

func timeToBarwidth(a time.Time) int {
	h, m, _ := a.Clock()
	return ((h * 60) + m) / MIN_DIVISOR
}

func rect(path draw2d.GraphicContext, x1, x2, y1, y2 float64) {
	path.MoveTo(x1, y1)
	path.LineTo(x2, y1)
	path.LineTo(x2, y2)
	path.LineTo(x1, y2)
	path.Close()
}

func arrow(path draw2d.GraphicContext, x, y, w, h float64, up bool) {
	var m float64 = 1
	if up {
		m = 3
	}

	path.MoveTo(x+w/2, y)
	path.LineTo(x+w, y+(h/4)*m)
	path.LineTo(x+w/2, y+h)
	path.LineTo(x, y+(h/4)*m)
	path.Close()
}

func drawUptimes(gc draw2d.GraphicContext) {
	y, m, d := curTime.Add(Day * -6).Date()
	sday := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	eday := sday.Add(Day - time.Second)

	gc.SetFontSize(40)

	// draw frame
	gc.SetStrokeColor(color.RGBA{0x00, 0x00, 0x00, 0xff})
	rect(gc, BAR_START_X, BAR_START_X+BAR_WIDTH, BAR_START_Y, BAR_START_Y+2+float64(7*BAR_HEIGHT))
	gc.Stroke()

	u := 0
	for i := 0; i < 7; i++ {
		gc.MoveTo(FONT_START_X, FONT_START_Y+float64(i*BAR_HEIGHT))
		gc.SetStrokeColor(color.RGBA{0x00, 0x00, 0x00, 0xff})
		gc.FillString(sday.Weekday().String()[0:2])

		upday := make([]w32uptime.Uptime, 0)
		for e := u; e < len(uptimes); e++ {
			if uptimes[e].Start.After(eday) {
				u = e - 1
				if u < 0 {
					u = 0
				}
				break
			}

			if isUptimeInRange(sday, eday, uptimes[e]) {
				upday = append(upday, uptimes[e])
			}
		}

		gc.SetFillColor(color.RGBA{0xE0, 0xA6, 0x2C, 0xFF})
		if len(upday) > 0 {
			for e := 0; e < len(upday); e++ {
				ps, pe := 0, BAR_WIDTH

				if upday[e].Start.After(sday) {
					ps = timeToBarwidth(upday[e].Start)
				}

				if upday[e].End.Before(eday) {
					pe = timeToBarwidth(upday[e].End)
				}

				rect(gc, float64(BAR_START_X+ps), float64(BAR_START_X+pe), BAR_START_Y+1+float64(i*BAR_HEIGHT), BAR_START_Y+1+float64((i+1)*BAR_HEIGHT))
				gc.Fill()
			}
		}

		if i != 6 {
			gc.SetStrokeColor(color.RGBA{0x00, 0x00, 0x00, 0xff})
			gc.MoveTo(BAR_START_X, BAR_START_Y+float64((i+1)*BAR_HEIGHT))
			gc.LineTo(BAR_START_X+BAR_WIDTH, BAR_START_Y+float64((i+1)*BAR_HEIGHT))
			gc.Close()
			gc.Stroke()
		}

		sday = sday.Add(Day)
		eday = sday.Add(Day - time.Second)
	}

	// middle line
	gc.SetStrokeColor(color.RGBA{0x00, 0x00, 0x00, 0xff})
	gc.MoveTo(BAR_START_X+BAR_WIDTH/2, BAR_START_Y+1)
	gc.LineTo(BAR_START_X+BAR_WIDTH/2, BAR_START_Y+1+float64(7*BAR_HEIGHT))
	gc.Close()
	gc.Stroke()
}

func drawarrowButtons(gc draw2d.GraphicContext) {
	for _, ab := range arrowButtons {
		ab.Paint(gc)
	}
}

func drawInfoarea(gc draw2d.GraphicContext) {
	gc.SetFontSize(20)
	gc.SetFillColor(color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
	gc.MoveTo(BAR_START_X, IA_START_Y)
	gc.FillString(curTime.Add(Day*-7).Format(DATE_FORMAT) + " - " + curTime.Format(DATE_FORMAT))
}

func drawBackground(gc draw2d.GraphicContext) {
	gc.SetFillColor(color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
	gc.Clear()
}

func redraw(gc draw2d.GraphicContext) {
	drawBackground(gc)
	drawUptimes(gc)
	drawarrowButtons(gc)
	drawInfoarea(gc)
}

func startGui() {
	w, err := wde.NewWindow(WIN_WIDTH, WIN_HEIGHT)
	if err != nil {
		panic(err)
	}
	w.SetTitle("winup")
	w.Show()

	arrowButtons = []ArrowButton{
		ArrowButton{
			SB_START_X, BAR_START_Y,
			SB_SIZE, SB_SIZE,
			true,
			func(ab ArrowButton) {
				curTime = curTime.Add(Day * -7)
			},
		},
		ArrowButton{
			SB_START_X, BAR_START_Y + 2 + 7*BAR_HEIGHT - SB_SIZE,
			SB_SIZE, SB_SIZE,
			false,
			func(ab ArrowButton) {
				curTime = curTime.Add(Day * 7)
			},
		},
	}

	gc := NewGraphicContext(w.Screen())
	curTime = time.Now()
	events := w.EventChan()
loop:
	for ei := range events {
		switch e := ei.(type) {
		case wde.CloseEvent:
			w.Close()
			break loop
		case wde.ResizeEvent:
			if e.Width == 0 || e.Height == 0 {
				continue
			}

			gc = NewGraphicContext(w.Screen())
			redraw(gc)
			w.FlushImage()
		case wde.MouseUpEvent:
			if e.Which != wde.LeftButton {
				continue
			}

			for _, ab := range arrowButtons {
				if ab.Contains(e.Where.X, e.Where.Y) {
					ab.Click()
					redraw(gc)
					w.FlushImage()
					break
				}
			}
		}
	}
	wde.Stop()
}

func main() {
	var err error
	uptimes, err = w32uptime.ReadAll()
	if err != nil {
		println(err.Error())
		return
	}

	go startGui()
	wde.Run()
}

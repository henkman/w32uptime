package main

import (
	"github.com/skelterjohn/go.wde"
	_ "github.com/skelterjohn/go.wde/init"
	"github.com/papplampe/w32uptime"
	"code.google.com/p/draw2d/draw2d"
	"image/color"
	"time"
)

const (
	Day = time.Duration(24) * time.Hour
	BAR_HEIGHT = 60
	BAR_WIDTH = 24 * 20
	MIN_DIVISOR = (24 * 60) / BAR_WIDTH
	FONT_START_Y = 60
	FONT_START_X = 15
	BAR_START_Y = FONT_START_Y - 45
	BAR_START_X = FONT_START_X + 95
	WIN_WIDTH = BAR_WIDTH + BAR_START_X + 10
	WIN_HEIGHT = BAR_START_Y + 7 * BAR_HEIGHT + 10
)

var (
	uptimes []w32uptime.Uptime
	curTime time.Time
)

func IsUptimeInRange(s, e time.Time, up w32uptime.Uptime) bool {
	return (up.Start.After(s) && up.Start.Before(e)) ||
		(up.End.After(s) && up.End.Before(e)) ||
		(up.Start.Before(s) && up.End.After(e))
}

func drawUptimes(gc draw2d.GraphicContext) {
	y, m, d := curTime.Add(Day * -6).Date()
	sday := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	eday := sday.Add(Day - time.Second)

	gc.SetFontSize(40)
	gc.SetFillColor(color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
	gc.Clear()
	
	// draw frame
	gc.SetFillColor(color.RGBA{0xE0, 0xA6, 0x2C, 0xFF})
	gc.MoveTo(BAR_START_X, BAR_START_Y + float64(7 * BAR_HEIGHT))
	gc.LineTo(BAR_START_X, BAR_START_Y)
	gc.LineTo(BAR_START_X + BAR_WIDTH, BAR_START_Y)
	gc.LineTo(BAR_START_X + BAR_WIDTH, BAR_START_Y + float64(7 * BAR_HEIGHT))
	gc.Close()
	gc.Stroke()
	
	u := 0
	for i := 0; i < 7; i++ {
		gc.MoveTo(FONT_START_X, FONT_START_Y + float64(i * BAR_HEIGHT))
		gc.SetStrokeColor(color.RGBA{0x00, 0x00, 0x00, 0xff})
		gc.FillString(sday.Weekday().String()[0:2])

		upday := make([]w32uptime.Uptime, 0)
		for e := u; e < len(uptimes); e++ {
			if uptimes[e].Start.After(eday) {
				u = e - 1
				break
			}
		
			if IsUptimeInRange(sday, eday, uptimes[e]) {
				upday = append(upday, uptimes[e])
			}
		}
		
		gc.SetFillColor(color.RGBA{0xFF, 0x00, 0x00, 0xff})
		if len(upday) > 0 {
			for e := 0; e < len(upday); e++ {
				ps, pe := 0, BAR_WIDTH
				
				if upday[e].Start.After(sday) {
					h, m, _ := upday[e].Start.Clock()
					m += h * 60
					ps = m / MIN_DIVISOR
				}
				
				if upday[e].End.Before(eday) {
					h, m, _ := upday[e].End.Clock()
					m += h * 60
					pe = m / MIN_DIVISOR
				}
				
				gc.MoveTo(float64(BAR_START_X + pe), BAR_START_Y + float64(i * BAR_HEIGHT))
				gc.LineTo(float64(BAR_START_X + ps), BAR_START_Y + float64(i * BAR_HEIGHT))
				gc.LineTo(float64(BAR_START_X + ps), BAR_START_Y + float64((i + 1) * BAR_HEIGHT))
				gc.LineTo(float64(BAR_START_X + pe), BAR_START_Y + float64((i + 1) * BAR_HEIGHT))
				gc.Close()
				gc.Fill()
			}
		}
		
		if i != 6 {
			gc.SetStrokeColor(color.RGBA{0x00, 0x00, 0x00, 0xff})
			gc.MoveTo(BAR_START_X, BAR_START_Y + float64((i + 1) * BAR_HEIGHT))
			gc.LineTo(BAR_START_X + BAR_WIDTH, BAR_START_Y + float64((i + 1) * BAR_HEIGHT))
			gc.Close()
			gc.Stroke()
		}
		
		sday = sday.Add(Day)
		eday = sday.Add(Day - time.Second)
	}
}	

func startGui() {
	w, err := wde.NewWindow(WIN_WIDTH, WIN_HEIGHT)
	if err != nil {
		panic(err)
	}
	w.SetTitle("winup")
	w.Show()
	
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
			drawUptimes(gc)
			w.FlushImage()
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

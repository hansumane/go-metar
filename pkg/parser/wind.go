package parser

import (
	"fmt"
	"io"
)

const (
	dirVariable = -1
	dirNA       = -2

	speedNA = -1
)

type Wind struct {
	Direction int
	Speed     int
	Gust      uint

	Variable struct {
		From uint
		To   uint
	}
}

func newWind(direction, speed, gust, units, varFrom, varTo string) Wind {
	var w Wind

	switch direction {
	case "VRB":
		w.Direction = dirVariable
	case "///":
		w.Direction = dirNA
	default:
		w.Direction = s2i(direction)
	}

	w.Variable.From = s2ui(varFrom)
	w.Variable.To = s2ui(varTo)

	if speed == "//" {
		w.Speed = speedNA
	} else {
		w.Speed = s2i(speed)
	}

	w.Gust = s2ui(gust)

	if units == "MPS" {
		if w.Speed != speedNA {
			w.Speed *= 2
		}
		w.Gust *= 2
	}

	return w
}

func fmtWindDir(direction int, speed int) (string, rune, string) {
	strSpeed := fmt.Sprintf("%-2d", speed)
	if speed == speedNA {
		strSpeed = "//"
	}

	switch direction {
	case dirVariable:
		return "VRB", ' ', strSpeed
	case dirNA:
		return "///", ' ', strSpeed
	}

	var dir rune
	if direction >= 360-22 || direction < 23 {
		dir = '↓'
	} else if direction >= 45-22 && direction <= 45+22 {
		dir = '↙'
	} else if direction >= 90-22 && direction <= 90+22 {
		dir = '←'
	} else if direction >= 135-22 && direction <= 135+22 {
		dir = '↖'
	} else if direction >= 180-22 && direction <= 180+22 {
		dir = '↑'
	} else if direction >= 225-22 && direction <= 225+22 {
		dir = '↗'
	} else if direction >= 270-22 && direction <= 270+22 {
		dir = '→'
	} else if direction >= 315-22 && direction <= 315+22 {
		dir = '↘'
	} else {
		dir = '?'
	}

	return fmt.Sprintf("%3d", direction), dir, strSpeed
}

func fmtDirectionSimple(direction int) string {
	switch direction {
	case dirVariable:
		return "VRB"
	default:
		return fmt.Sprintf("%03d", uint(direction))
	}
}

func (w Wind) Format(s fmt.State, v rune) {
	switch v {
	case 'v':
		if s.Flag('+') {
			io.WriteString(s, fmt.Sprintf(
				"Wind{Direction:%v Speed:%v, Gust:%v Variable{From:%v To:%v}}",
				w.Direction, w.Speed, w.Gust, w.Variable.From, w.Variable.To,
			))
		} else {
			dir, _, _ := fmtWindDir(w.Direction, w.Speed)
			io.WriteString(s, "Wind{")
			if w.Variable.From != w.Variable.To {
				io.WriteString(s, fmt.Sprintf(
					"%03d-%v-%03d",
					w.Variable.From,
					dir,
					w.Variable.To,
				))
			} else {
				io.WriteString(s, fmtDirectionSimple(w.Direction))
			}
			if w.Speed != speedNA {
				io.WriteString(s, fmt.Sprintf(" @ %d", w.Speed))
				if w.Gust != 0 {
					io.WriteString(s, fmt.Sprintf("-%d", w.Gust))
				}
				io.WriteString(s, " kts")
			}
			io.WriteString(s, "}")
		}
	case 's':
		dir, arrow, speed := fmtWindDir(w.Direction, w.Speed)
		io.WriteString(s, fmt.Sprintf("%s %c %s", dir, arrow, speed))
	default:
		io.WriteString(s, "Wind{}")
	}
}

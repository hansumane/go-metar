package parser

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

const (
	metarParsedLen = 37

	windRegex    = ` ([\d\/]{3}|VRB)([\d\/]{2})(?:G(\d{2}))?(MPS|KT)`
	wxs          = `MI|PR|BC|DR|BL|SH|TS|FZ|DZ|RA|SN|SG|GS|GR|PL|IC|UP|FG|BR|HZ|VA|DU|FU|SA|PY|SQ|PO|DS|SS|FC`
	weatherRegex = `((?: (?:VC|[\-\+])?(?:`+wxs+`)+)*| \/{2})`
	cloudRegex   = `((?: (?:(?:FEW|SCT|BKN|OVC|VV|\/{3})(?:\d{3}(?:TCU|CB|\/{3})?|\/{3})|CAVOK|SKC|NCD|CLR|NSC))*)`
)

const metarRegex = `^` +
	// [1] => type
	`(METAR|SPECI)` +
	// [2] => icao id
	` ([A-Z]{4})` +
	// [3,4,5] => UTC day, hour, minutes
	` (\d{2})(\d{2})(\d{2})Z` +
	// [6] => modifier flag
	`((?: (?:AUTO|COR))*)` +
	// [7,8,9?,10] => wind direction, speed, gusts, units
	windRegex +
	// [11?,12?] => variable wind directions
	`(?: (\d{3})V(\d{3}))?` +
	// [13?] => visibility string
	`(?: ([MP]?[\d\/]{4}(?: [MP]?\d{4}(?:N|NE|E|SE|S|SW|W|NW))*|(?:\d{1,2} )?[MP]?\d{1,2}(?:\/(?:2|4|8|16|32))?SM))?` +
	// [14?] => RVR string
	`((?: R\d{1,2}[LCR]?\/[MP]?\d{4}[DNU]?(?:V[MP]?\d{4}[DNU]?)?(?:FT)?)*)` +
	// [15?] => weather string
	weatherRegex +
	// [16?] => cloud string
	cloudRegex +
	// [17?,18,19?,20] => temperature and dew point
	` (?:(M?)(\d{2})\/(M?)(\d{2}))` +
	// [21, 22] => pressure
	` (?:(Q|A)(\d{4}))` +
	// [23?] => recent most significant weather
	`(?: RE((?:`+wxs+`)+))?` +
	// [24?] => windshear string
	`((?: WS R\d{1,2}[LCR]?)*)` +
	// [25?] => runway friction string
	`((?: R\d{1,2}[LCR]?\/(?:CLRD|\d{4})\d{2})*)` +
	// [26]? => trend group
	`(?: (TEMPO|BECMG)` +
		// [27?,28?,29?] => from, until, at
		`(?:(?: FM(\d{4}))? TL(\d{4})| AT(\d{4}))?` +
		// [30,31,32?,33]? => wind direction, speed, gusts, units
		`(?:`+windRegex+`)?` +
		// [34?] => visibility (simple)
		`(?: (\d{4}))?` +
		// [35?] => weather string
		weatherRegex +
		// [36?] => cloud string
		cloudRegex +
	`)?`

type Metar struct {
	Raw string

	Type string
	ID   string

	DateTime struct {
		Day    uint
		Hour   uint
		Minute uint
	}

	Modifiers []string

	Wind Wind

	// TODO: Visibility
	// TODO: RVRs

	Weather string // TODO: Parse Weather
	Clouds  string // TODO: Parse Clouds

	Temperature int
	DewPoint    int

	Pressure struct {
		Unit  string
		Value uint
	}

	ReWeather string // TODO: Parse Weather

	// TODO: Windshear
	// TODO: Runway Friction

	Trend struct {
		Type    string
		Weather string // TODO: Parse Weather
		Clouds  string // TODO: Parse Clouds
	}
}

func NewMetar(raw string) Metar {
	return Metar{
		Raw: strings.Join(strings.Fields(raw), " "),
	}
}

func (m *Metar) Parse() error {
	re := regexp.MustCompile(metarRegex)
	parsed := re.FindStringSubmatch(m.Raw)

	if len(parsed) != metarParsedLen {
		return fmt.Errorf("bad metar: %q", m.Raw)
	}

	m.Type = parsed[1]
	m.ID = parsed[2]
	m.DateTime.Day = s2ui(parsed[3])
	m.DateTime.Hour = s2ui(parsed[4])
	m.DateTime.Minute = s2ui(parsed[5])

	m.Modifiers = unique(strings.Fields(parsed[6]))

	m.Wind = newWind(parsed[7], parsed[8], parsed[9], parsed[10], parsed[11], parsed[12])

	m.Weather = parsed[15]
	m.Clouds = parsed[16]

	m.Temperature = s2i(parsed[18])
	if parsed[17] == "M" {
		m.Temperature *= -1
	}
	m.DewPoint = s2i(parsed[20])
	if parsed[19] == "M" {
		m.DewPoint *= -1
	}

	m.Pressure.Unit = parsed[21]
	m.Pressure.Value = s2ui(parsed[22])

	m.ReWeather = parsed[23]

	m.Trend.Type = parsed[26]
	m.Trend.Weather = parsed[35]
	m.Trend.Clouds = parsed[36]

	return nil
}

func fmtWxClouds(main, trend, recent, trendType string) string {
	if main == " CAVOK" {
		main = ""
	}

	if recent != "" {
		main += fmt.Sprintf(" RE:%s", recent)
	}

	for part := range strings.FieldsSeq(trend) {
		main += fmt.Sprintf(" %c:%s", trendType[0], part)
	}

	return main
}

func (m Metar) Format(s fmt.State, v rune) {
	switch v {
	case 'v':
		if !s.Flag('+') {
			io.WriteString(s, fmt.Sprintf("Metar{%s}\n", m.ID))
			return
		}

		weather := fmtWxClouds(m.Weather, m.Trend.Weather, m.ReWeather, m.Trend.Type)
		clouds := fmtWxClouds(m.Clouds, m.Trend.Clouds, "", m.Trend.Type)

		humidity := calcHumidity(m.Temperature, m.DewPoint)
		if humidity >= 99 {
			humidity = 99
		}

		reWeather := fmt.Sprintf("RE%s", m.ReWeather)
		if reWeather == "RE" {
			reWeather = ""
		}

		io.WriteString(s, fmt.Sprintf(
			"Metar{\n"+
				"\tType:%s\n"+
				"\tId:%s\n"+
				"\tDateTime:%+v\n"+
				"\tModifiers:%v\n"+
				"\tWind:%v\n"+
				"\tWeather:%v\n"+
				"\tClouds:%v\n"+
				"\tTemperature: %d/%d (%.0f%%)\n"+
				"\tPressure:{%s %04d}\n"+
			"}",
			m.Type,
			m.ID,
			m.DateTime,
			m.Modifiers,
			m.Wind,
			weather,
			clouds,
			m.Temperature, m.DewPoint, humidity,
			m.Pressure.Unit, m.Pressure.Value,
		))
	case 's':
		mod := ' '
		if len(m.Modifiers) == 1 {
			mod = []rune(m.Modifiers[0])[0]
		} else if len(m.Modifiers) > 1 {
			mod = '?'
		}

		io.WriteString(s, fmt.Sprintf(
			"[%c%c] %s day %02d at %02d:%02d UTC: "+
				"%s %+3d/%-+3d %2.0f%% "+
				"%s%04d"+
				"%s%s",
			m.Type[0], mod, m.ID, m.DateTime.Day, m.DateTime.Hour, m.DateTime.Minute,
			m.Wind, m.Temperature, m.DewPoint, calcHumidity(m.Temperature, m.DewPoint),
			m.Pressure.Unit, m.Pressure.Value,
			fmtWxClouds(m.Weather, m.Trend.Weather, m.ReWeather, m.Trend.Type),
			fmtWxClouds(m.Clouds, m.Trend.Clouds, "", m.Trend.Type),
		))
	default:
		io.WriteString(s, "Metar{}")
	}
}

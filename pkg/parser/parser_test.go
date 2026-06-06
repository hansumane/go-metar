package parser

import (
	"testing"
)

func TestParser(t *testing.T) {
	metars := [...]string{
		"METAR ZZZZ 070051Z 32014G22KT P6SM -TSRAGR BR SCT040TCU BKN070CB 22/17 A2982 " +
		"RMK AO2 PK WND 30052/0003 WSHFT 2349 RAB2359 TSB2359E47 SLP097 P0000 T02170172 $",

		"METAR ZZZZ 070100Z AUTO /////MPS //// // ////// 15/12 Q1012 RMK SLP123 T01510122",
	}
	for _, raw := range metars {
		m := NewMetar(raw)
		if err := m.Parse(); err != nil {
			t.Errorf("could not parse: %s", err)
			t.FailNow()
		} else {
			t.Logf("parsed: %s", m)
		}
	}
}

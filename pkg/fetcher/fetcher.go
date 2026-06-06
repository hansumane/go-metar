package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type metar struct {
	ID  string `json:"icaoId"`
	Raw string `json:"rawOb"`
}

func FetchMetars(airports []string) ([]string, error) {
	if len(airports) == 0 {
		return nil, fmt.Errorf("no airports")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	url := "https://aviationweather.gov/api/data/metar?format=json&ids=" + strings.Join(airports, ",")
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if body, err := io.ReadAll(resp.Body); err != nil {
			return nil, fmt.Errorf("bad status: %s", resp.Status)
		} else {
			return nil, fmt.Errorf("bad status: %s - %s", resp.Status, string(body))
		}
	}

	var metars []metar
	if err := json.NewDecoder(resp.Body).Decode(&metars); err != nil {
		return nil, err
	}

	metarsMap := make(map[string]string, len(metars))
	for _, m := range metars {
		metarsMap[m.ID] = m.Raw
	}

	var out []string
	for _, icao := range airports {
		if raw, ok := metarsMap[icao]; ok {
			out = append(out, raw)
		}
	}

	return out, nil
}

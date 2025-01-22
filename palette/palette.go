package palette

import (
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"strings"
	"sync"
)

type PaletteInfo struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type Palette map[string][]string

var (
	palettes Palette
	mu       sync.RWMutex
)

func LoadPalettes(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error when opening the palette file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&palettes); err != nil {
		return fmt.Errorf("error decoding JSON palettes: %v", err)
	}

	return nil
}

func GetPaletteInfos() []PaletteInfo {
	mu.RLock()
	defer mu.RUnlock()

	paletteInfos := make([]PaletteInfo, 0, len(palettes))
	for name, colors := range palettes {
		paletteInfos = append(paletteInfos, PaletteInfo{Name: name, Count: len(colors)})
	}
	return paletteInfos
}

func GetPaletteHex(name string) ([]string, bool) {
	mu.RLock()
	defer mu.RUnlock()

	colors, exists := palettes[name]
	return colors, exists
}

func GetDefaultPaletteHex() []string {
	mu.RLock()
	defer mu.RUnlock()

	return palettes["default"]
}

func ParsePalette(paletteHex []string) ([]color.Color, error) {
	return parsePalette(paletteHex)
}

func parsePalette(paletteHex []string) ([]color.Color, error) {
	palette := make([]color.Color, 0, len(paletteHex))
	for _, hex := range paletteHex {
		c, err := parseHexColor(hex)
		if err != nil {
			return nil, fmt.Errorf("color parsing error %s: %v", hex, err)
		}
		palette = append(palette, c)
	}
	return palette, nil
}

func parseHexColor(s string) (color.Color, error) {
	c := color.RGBA{A: 0xff}
	if strings.HasPrefix(s, "#") {
		s = s[1:]
	}
	var err error
	switch len(s) {
	case 6:
		_, err = fmt.Sscanf(s, "%02x%02x%02x", &c.R, &c.G, &c.B)
	case 3:
		var r, g, b uint8
		_, err = fmt.Sscanf(s, "%1x%1x%1x", &r, &g, &b)
		c.R = r * 17
		c.G = g * 17
		c.B = b * 17
	default:
		err = fmt.Errorf("invalid string length")
	}
	return c, err
}

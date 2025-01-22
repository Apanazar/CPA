package handlers

import (
	"encoding/json"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"strconv"
	"strings"

	"cpa/imageprocessing"
	"cpa/palette"

	"github.com/disintegration/imaging"
)

func LoadPalettes(path string) error {
	return palette.LoadPalettes(path)
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "The method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.ServeFile(w, r, "templates/index.html")
}

func PalettesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "The method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	paletteInfos := palette.GetPaletteInfos()

	paletteInfos = append([]palette.PaletteInfo{
		{Name: "default", Count: 0},
		{Name: "original", Count: 20},
	}, paletteInfos...)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(paletteInfos); err != nil {
		http.Error(w, "Error when encoding JSON", http.StatusInternalServerError)
		return
	}
}

func ExtractPaletteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "The method is not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Error when parsing the form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error when receiving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		http.Error(w, "Error when decoding the image", http.StatusBadRequest)
		return
	}

	paletteColors, err := imageprocessing.KMeans(img, 20, 10)
	if err != nil {
		http.Error(w, "Error when executing K-Means", http.StatusInternalServerError)
		return
	}

	paletteHex := imageprocessing.ColorsToHex(paletteColors)

	response := struct {
		Palette []string            `json:"palette"`
		Info    palette.PaletteInfo `json:"info"`
	}{
		Palette: paletteHex,
		Info: palette.PaletteInfo{
			Name:  "original",
			Count: len(paletteHex),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error when encoding JSON", http.StatusInternalServerError)
		return
	}
}

func ProcessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "The method is not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 50<<20)

	if err := r.ParseMultipartForm(50 << 20); err != nil {
		http.Error(w, "Error when parsing the form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error when receiving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		http.Error(w, "Error when decoding the image", http.StatusBadRequest)
		return
	}

	blockSize, err := strconv.Atoi(r.FormValue("blocksize"))
	if err != nil || blockSize < 2 {
		blockSize = 32
	}

	paletteName := r.FormValue("palette")
	if paletteName == "" {
		paletteName = "default"
	}

	useAllColors, err := strconv.Atoi(r.FormValue("useAllColors"))
	if err != nil || useAllColors < 1 {
		useAllColors = 1
	}

	brightness, err := strconv.Atoi(r.FormValue("brightness"))
	if err != nil {
		brightness = 0
	}

	contrast, err := strconv.Atoi(r.FormValue("contrast"))
	if err != nil {
		contrast = 0
	}

	saturation, err := strconv.Atoi(r.FormValue("saturation"))
	if err != nil {
		saturation = 0
	}

	contourStr := r.FormValue("contour")
	contour := contourStr == "on"

	resolutionStr := r.FormValue("resolution")
	desiredRes, err := strconv.Atoi(resolutionStr)
	if err != nil {
		desiredRes = 0
	}

	origBounds := img.Bounds()
	origW, origH := origBounds.Dx(), origBounds.Dy()

	var adjusted image.Image = img
	if brightness != 0 {
		adjusted = imageprocessing.AdjustBrightness(adjusted, float64(brightness))
	}
	if contrast != 0 {
		adjusted = imageprocessing.AdjustContrast(adjusted, float64(contrast))
	}
	if saturation != 0 {
		adjusted = imageprocessing.AdjustSaturation(adjusted, float64(saturation))
	}

	var finalImg image.Image
	if desiredRes > 0 {
		downscaled := imaging.Resize(adjusted, desiredRes, desiredRes, imaging.NearestNeighbor)
		upscaled := imaging.Resize(downscaled, origW, origH, imaging.NearestNeighbor)
		finalImg = upscaled
	} else {
		pixelated := imageprocessing.PixelateImageCustom(adjusted, blockSize)
		finalImg = pixelated
	}

	switch strings.ToLower(paletteName) {
	case "default":
		// ...
	case "original":
		paletteDataStr := r.FormValue("palette_data")
		if paletteDataStr == "" {
			http.Error(w, "The 'original' palette requires palette data (palette_data)", http.StatusBadRequest)
			return
		}
		var paletteHex []string
		if err := json.Unmarshal([]byte(paletteDataStr), &paletteHex); err != nil {
			http.Error(w, "Error when decoding the palette (original)", http.StatusBadRequest)
			return
		}
		parsedPalette, err := palette.ParsePalette(paletteHex)
		if err != nil {
			http.Error(w, "Error when parsing the palette (original)", http.StatusInternalServerError)
			return
		}
		if useAllColors > len(parsedPalette) {
			useAllColors = len(parsedPalette)
		}
		applied, err := imageprocessing.ApplyPalette(finalImg, parsedPalette, blockSize, useAllColors)
		if err != nil {
			http.Error(w, "Error when applying the palette (original)", http.StatusInternalServerError)
			return
		}
		finalImg = applied
	default:
		paletteHex, exists := palette.GetPaletteHex(paletteName)
		if !exists {
			paletteHex = palette.GetDefaultPaletteHex()
		}
		parsedPalette, err := palette.ParsePalette(paletteHex)
		if err != nil {
			http.Error(w, "Error when parsing the palette "+paletteName, http.StatusInternalServerError)
			return
		}
		if useAllColors > len(parsedPalette) {
			useAllColors = len(parsedPalette)
		}
		applied, err := imageprocessing.ApplyPalette(finalImg, parsedPalette, blockSize, useAllColors)
		if err != nil {
			http.Error(w, "Error when applying the palette "+paletteName, http.StatusInternalServerError)
			return
		}
		finalImg = applied
	}

	if contour {
		contourImg := imageprocessing.ApplyContour(finalImg, 100.0)
		grayContourImg, ok := contourImg.(*image.Gray)
		if !ok {
			http.Error(w, "Error: ApplyContour did not return *image.Gray", http.StatusInternalServerError)
			return
		}
		finalImg = imageprocessing.ApplyBlackContours(finalImg, grayContourImg)
	}

	if strings.ToLower(format) == "jpeg" || strings.ToLower(format) == "jpg" {
		w.Header().Set("Content-Type", "image/jpeg")
		if err := jpeg.Encode(w, finalImg, nil); err != nil {
			http.Error(w, "Error encoding the result (JPEG)", http.StatusInternalServerError)
			return
		}
	} else {
		w.Header().Set("Content-Type", "image/png")
		if err := png.Encode(w, finalImg); err != nil {
			http.Error(w, "Error encoding the result (PNG)", http.StatusInternalServerError)
			return
		}
	}
}

package imageprocessing

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sort"

	"math/rand"

	"github.com/disintegration/imaging"
	"golang.org/x/image/draw"
)

type Cluster struct {
	Centroid color.Color
	Points   []color.Color
}

func AdjustBrightness(img image.Image, brightness float64) image.Image {
	return imaging.AdjustBrightness(img, brightness)
}

func AdjustContrast(img image.Image, contrast float64) image.Image {
	return imaging.AdjustContrast(img, contrast)
}

func AdjustSaturation(img image.Image, saturation float64) image.Image {
	return imaging.AdjustSaturation(img, saturation)
}

func PixelateImageCustom(img image.Image, blockSize int) image.Image {
	return pixelateImageCustom(img, blockSize)
}

func pixelateImageCustom(img image.Image, blockSize int) image.Image {
	bounds := img.Bounds()
	pixelated := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y += blockSize {
		for x := bounds.Min.X; x < bounds.Max.X; x += blockSize {
			endY := y + blockSize
			if endY > bounds.Max.Y {
				endY = bounds.Max.Y
			}
			endX := x + blockSize
			if endX > bounds.Max.X {
				endX = bounds.Max.X
			}

			avgColor := averageColor(img, x, y, endX, endY)

			for by := y; by < endY; by++ {
				for bx := x; bx < endX; bx++ {
					pixelated.Set(bx, by, avgColor)
				}
			}
		}
	}

	return pixelated
}

func ApplyPalette(img image.Image, palette []color.Color, blockSize int, useAllColors int) (image.Image, error) {
	return applyPalette(img, palette, blockSize, useAllColors)
}

func applyPalette(img image.Image, palette []color.Color, blockSize int, useAllColors int) (image.Image, error) {
	bounds := img.Bounds()
	finalImg := image.NewRGBA(bounds)

	totalColors := len(palette)
	if totalColors == 0 {
		return nil, fmt.Errorf("the palette is empty")
	}

	if useAllColors > totalColors {
		useAllColors = totalColors
	}
	if useAllColors < 1 {
		useAllColors = 1
	}

	topColors := getTopNColors(palette, useAllColors)

	for y := bounds.Min.Y; y < bounds.Max.Y; y += blockSize {
		for x := bounds.Min.X; x < bounds.Max.X; x += blockSize {
			endY := y + blockSize
			if endY > bounds.Max.Y {
				endY = bounds.Max.Y
			}
			endX := x + blockSize
			if endX > bounds.Max.X {
				endX = bounds.Max.X
			}

			avgColor := img.At(x, y)

			selectedColor := getClosestColor(avgColor, topColors)

			for by := y; by < endY; by++ {
				for bx := x; bx < endX; bx++ {
					finalImg.Set(bx, by, selectedColor)
				}
			}
		}
	}

	return finalImg, nil
}

func ApplyContour(img image.Image, threshold float64) image.Image {
	return applyContour(img, threshold)
}

func applyContour(img image.Image, threshold float64) image.Image {
	grayImg := convertToGray(img)
	sobelImg := applySobel(grayImg)
	thresholdedImg := thresholdImage(sobelImg, threshold)
	return thresholdedImg
}

func ApplyBlackContours(originalImg image.Image, contourImg *image.Gray) image.Image {
	rgbaImg := image.NewRGBA(originalImg.Bounds())
	draw.Draw(rgbaImg, rgbaImg.Bounds(), originalImg, image.Point{}, draw.Src)

	bounds := contourImg.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if contourImg.GrayAt(x, y).Y == 0 {
				rgbaImg.Set(x, y, color.Black)
			}
		}
	}

	return rgbaImg
}

func KMeans(img image.Image, k int, maxIterations int) ([]color.Color, error) {
	return kMeans(img, k, maxIterations)
}

func kMeans(img image.Image, k int, maxIterations int) ([]color.Color, error) {
	if k <= 0 {
		return nil, fmt.Errorf("the number of clusters must be positive")
	}

	clusters := make([]Cluster, k)
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	for i := 0; i < k; i++ {
		x := rand.Intn(width)
		y := rand.Intn(height)
		clusters[i].Centroid = img.At(x, y)
	}

	for iter := 0; iter < maxIterations; iter++ {
		for i := 0; i < k; i++ {
			clusters[i].Points = []color.Color{}
		}

		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				currentColor := img.At(x, y)
				minDist := math.MaxFloat64
				clusterIndex := 0
				for i, cluster := range clusters {
					dist := colorDistanceMetric(currentColor, cluster.Centroid)
					if dist < minDist {
						minDist = dist
						clusterIndex = i
					}
				}
				clusters[clusterIndex].Points = append(clusters[clusterIndex].Points, currentColor)
			}
		}

		converged := true
		for i := 0; i < k; i++ {
			newCentroid := calculateCentroid(clusters[i].Points)
			if colorDistanceMetric(newCentroid, clusters[i].Centroid) > 1.0 {
				converged = false
				clusters[i].Centroid = newCentroid
			}
		}

		if converged {
			fmt.Printf("Convergence is achieved in the iteration %d\n", iter)
			break
		}
	}

	palette := make([]color.Color, k)
	for i, cluster := range clusters {
		palette[i] = cluster.Centroid
	}

	return palette, nil
}

func ColorsToHex(palette []color.Color) []string {
	return colorsToHex(palette)
}

func colorsToHex(palette []color.Color) []string {
	hexPalette := make([]string, len(palette))
	for i, c := range palette {
		rgba := color.RGBAModel.Convert(c).(color.RGBA)
		hexPalette[i] = fmt.Sprintf("#%02x%02x%02x", rgba.R, rgba.G, rgba.B)
	}
	return hexPalette
}

func averageColor(img image.Image, startX, startY, endX, endY int) color.Color {
	var r, g, b, a uint32
	count := uint32(0)

	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			cr, cg, cb, ca := img.At(x, y).RGBA()
			r += cr
			g += cg
			b += cb
			a += ca
			count++
		}
	}

	if count == 0 {
		return color.Black
	}

	return color.RGBA{
		R: uint8((r / count) >> 8),
		G: uint8((g / count) >> 8),
		B: uint8((b / count) >> 8),
		A: uint8((a / count) >> 8),
	}
}

func getTopNColors(palette []color.Color, n int) []color.Color {
	var r, g, b, a uint32
	count := 0
	for _, c := range palette {
		cr, cg, cb, ca := c.RGBA()
		r += cr
		g += cg
		b += cb
		a += ca
		count++
	}
	if count == 0 {
		return []color.Color{color.Black}
	}
	avgColor := color.RGBA{
		R: uint8((r / uint32(count)) >> 8),
		G: uint8((g / uint32(count)) >> 8),
		B: uint8((b / uint32(count)) >> 8),
		A: uint8((a / uint32(count)) >> 8),
	}

	sortedPalette := sortedPaletteByDistance(avgColor, palette)

	if n > len(sortedPalette) {
		n = len(sortedPalette)
	}
	return sortedPalette[:n]
}

func getClosestColor(target color.Color, palette []color.Color) color.Color {
	minDist := math.MaxFloat64
	var closest color.Color
	for _, c := range palette {
		dist := colorDistanceMetric(target, c)
		if dist < minDist {
			minDist = dist
			closest = c
		}
	}
	return closest
}

func sortedPaletteByDistance(target color.Color, palette []color.Color) []color.Color {
	type colorDistance struct {
		col  color.Color
		dist float64
	}
	distances := make([]colorDistance, len(palette))
	for i, pc := range palette {
		distances[i].col = pc
		distances[i].dist = colorDistanceMetric(target, pc)
	}
	sort.Slice(distances, func(i, j int) bool {
		return distances[i].dist < distances[j].dist
	})
	sorted := make([]color.Color, len(palette))
	for i, cd := range distances {
		sorted[i] = cd.col
	}
	return sorted
}

func colorDistanceMetric(c1, c2 color.Color) float64 {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()

	dr := float64(r1>>8) - float64(r2>>8)
	dg := float64(g1>>8) - float64(g2>>8)
	db := float64(b1>>8) - float64(b2>>8)

	return math.Sqrt(dr*dr + dg*dg + db*db)
}

func applySobel(grayImg *image.Gray) *image.Gray {
	bounds := grayImg.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	sobelImg := image.NewGray(bounds)

	gx := [3][3]int{
		{-1, 0, 1},
		{-2, 0, 2},
		{-1, 0, 1},
	}

	gy := [3][3]int{
		{-1, -2, -1},
		{0, 0, 0},
		{1, 2, 1},
	}

	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			var sumX, sumY int
			for ky := -1; ky <= 1; ky++ {
				for kx := -1; kx <= 1; kx++ {
					pixel := grayImg.GrayAt(x+kx, y+ky).Y
					sumX += int(pixel) * gx[ky+1][kx+1]
					sumY += int(pixel) * gy[ky+1][kx+1]
				}
			}

			magnitude := math.Sqrt(float64(sumX*sumX + sumY*sumY))
			if magnitude > 255 {
				magnitude = 255
			}
			if magnitude < 0 {
				magnitude = 0
			}

			sobelImg.SetGray(x, y, color.Gray{Y: uint8(magnitude)})
		}
	}

	return sobelImg
}

func thresholdImage(sobelImg *image.Gray, threshold float64) *image.Gray {
	bounds := sobelImg.Bounds()
	thresholdedImg := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixel := sobelImg.GrayAt(x, y).Y
			if float64(pixel) > threshold {
				thresholdedImg.SetGray(x, y, color.Gray{Y: 0})
			} else {
				thresholdedImg.SetGray(x, y, color.Gray{Y: 255})
			}
		}
	}

	return thresholdedImg
}

func convertToGray(img image.Image) *image.Gray {
	bounds := img.Bounds()
	grayImg := image.NewGray(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalColor := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
			grayImg.SetGray(x, y, originalColor)
		}
	}
	return grayImg
}

func calculateCentroid(points []color.Color) color.Color {
	var r, g, b, a uint32
	for _, c := range points {
		cr, cg, cb, ca := c.RGBA()
		r += cr
		g += cg
		b += cb
		a += ca
	}
	n := uint32(len(points))
	if n == 0 {
		return color.Black
	}
	return color.RGBA{
		R: uint8((r / n) >> 8),
		G: uint8((g / n) >> 8),
		B: uint8((b / n) >> 8),
		A: uint8((a / n) >> 8),
	}
}

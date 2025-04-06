package pogoicon

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"math"
	"path"

	"golang.org/x/image/draw"
)

func Generate(assets fs.FS, event EventConfig) (io.Reader, error) {
	background, err := assets.Open(path.Join("assets/backgrounds", event.Background))
	if err != nil {
		return nil, fmt.Errorf("failed to open background asset: %w", err)
	}
	defer background.Close()

	overlays := make([]Overlay, 0, len(event.Overlays))
	for _, overlay := range event.Overlays {
		img, err := assets.Open(path.Join("assets/icons", overlay.Image))
		if err != nil {
			return nil, fmt.Errorf("failed to open overlay asset: %w", err)
		}
		defer img.Close()

		overlays = append(overlays, Overlay{
			Image:    img,
			ScaleX:   overlay.ScaleX,
			ScaleY:   overlay.ScaleY,
			Position: overlay.Position,
			OffsetX:  overlay.OffsetX,
			OffsetY:  overlay.OffsetY,
			FlipX:    overlay.FlipX,
			FlipY:    overlay.FlipY,
			Rotate:   overlay.Rotate,
		})
	}

	return generate(background, overlays)
}

func generate(background io.Reader, overlays []Overlay) (io.Reader, error) {
	backgroundImage, _, err := image.Decode(background)
	if err != nil {
		return nil, fmt.Errorf("failed to decode background image: %w", err)
	}

	newImage := image.NewRGBA(backgroundImage.Bounds())
	draw.Draw(newImage, newImage.Bounds(), backgroundImage, image.Point{}, draw.Src)

	for _, overlay := range overlays {
		if err = applyOverlay(newImage, overlay); err != nil {
			return nil, fmt.Errorf("failed to overlay template: %w", err)
		}
	}

	buf := new(bytes.Buffer)
	if err = png.Encode(buf, newImage); err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}
	return bytes.NewReader(buf.Bytes()), nil
}

func applyOverlay(baseImg *image.RGBA, overlay Overlay) error {
	img, _, err := image.Decode(overlay.Image)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	var (
		scaledWidth  int
		scaledHeight int
	)
	if overlay.ScaleX != 0 {
		scaledWidth = int(float64(baseImg.Bounds().Dx()) * overlay.ScaleX)
	} else {
		scaledWidth = img.Bounds().Dx()
	}
	if overlay.ScaleY != 0 {
		scaledHeight = int(float64(baseImg.Bounds().Dy()) * overlay.ScaleY)
	} else {
		scaledHeight = img.Bounds().Dy()
	}
	img = resizeImage(img, float64(scaledWidth)/float64(img.Bounds().Dx()), float64(scaledHeight)/float64(img.Bounds().Dy()))
	img = flipImage(img, overlay.FlipX, overlay.FlipY)
	img = rotateImage(img, overlay.Rotate)

	bounds := img.Bounds()
	var (
		offsetX int
		offsetY int
	)
	switch overlay.Position {
	case PositionTop:
		offsetY = 0
	case PositionTopLeft:
		offsetX = 0
		offsetY = 0
	case PositionTopRight:
		offsetX = baseImg.Bounds().Dx() - bounds.Dx()
		offsetY = 0
	case PositionBottom:
		offsetY = baseImg.Bounds().Dy() - bounds.Dy()
	case PositionBottomLeft:
		offsetX = 0
		offsetY = baseImg.Bounds().Dy() - bounds.Dy()
	case PositionBottomRight:
		offsetX = baseImg.Bounds().Dx() - bounds.Dx()
		offsetY = baseImg.Bounds().Dy() - bounds.Dy()
	case PositionCenter:
		offsetX = (baseImg.Bounds().Dx() - bounds.Dx()) / 2
		offsetY = (baseImg.Bounds().Dy() - bounds.Dy()) / 2
	case PositionLeft:
		offsetX = 0
		offsetY = (baseImg.Bounds().Dy() - bounds.Dy()) / 2
	case PositionRight:
		offsetX = baseImg.Bounds().Dx() - bounds.Dx()
		offsetY = (baseImg.Bounds().Dy() - bounds.Dy()) / 2
	default:
		return fmt.Errorf("invalid position: %s", overlay.Position)
	}
	offsetX += overlay.OffsetX
	offsetY += overlay.OffsetY

	draw.Draw(baseImg, baseImg.Bounds(), img, image.Point{
		X: offsetX,
		Y: offsetY,
	}, draw.Over)

	return nil
}

func resizeImage(img image.Image, scaleX float64, scaleY float64) image.Image {
	bounds := img.Bounds()
	newWidth := int(float64(bounds.Dx()) * scaleX)
	newHeight := int(float64(bounds.Dy()) * scaleY)

	resizedImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.BiLinear.Scale(resizedImg, resizedImg.Bounds(), img, bounds, draw.Over, nil)
	return resizedImg
}

func flipImage(img image.Image, flipX bool, flipY bool) image.Image {
	if flipX {
		img = flipXImage(img)
	}
	if flipY {
		img = flipYImage(img)
	}
	return img
}

func flipXImage(img image.Image) image.Image {
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			newImg.Set(bounds.Dx()-1-x, y, img.At(x, y))
		}
	}
	return newImg
}

func flipYImage(img image.Image) image.Image {
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			newImg.Set(x, bounds.Dy()-1-y, img.At(x, y))
		}
	}
	return newImg
}

func rotateImage(img image.Image, angle float64) image.Image {
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)
	angle = angle * (3.141592653589793 / 180.0) // Convert degrees to radians
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			newX := int(float64(x)*math.Cos(angle) - float64(y)*math.Sin(angle))
			newY := int(float64(x)*math.Sin(angle) + float64(y)*math.Cos(angle))
			newImg.Set(newX, newY, img.At(x, y))
		}
	}
	return newImg
}

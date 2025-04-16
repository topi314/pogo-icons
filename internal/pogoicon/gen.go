package pogoicon

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"math"

	"golang.org/x/image/draw"
)

func Generate(overlays []Overlay) (io.Reader, error) {
	var newImage *image.RGBA
	for i, overlay := range overlays {
		img, _, err := image.Decode(overlay.Image)
		if err != nil {
			return nil, fmt.Errorf("failed to decode image: %w", err)
		}
		if i == 0 {
			newImage = image.NewRGBA(img.Bounds())
		}

		if err = applyOverlay(newImage, img, overlay); err != nil {
			return nil, fmt.Errorf("failed to overlay template: %w", err)
		}
	}

	buf := new(bytes.Buffer)
	if err := png.Encode(buf, newImage); err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}
	return bytes.NewReader(buf.Bytes()), nil
}

func applyOverlay(baseImg *image.RGBA, img image.Image, overlay Overlay) error {
	img = resizeImage(baseImg, img, overlay.ScaleX, overlay.ScaleY)
	img = flipImage(img, overlay.FlipX, overlay.FlipY)
	img = rotateImage(img, overlay.Rotate)

	bounds := img.Bounds()
	baseBounds := baseImg.Bounds()
	var (
		offsetX int
		offsetY int
	)
	switch overlay.Position {
	case PositionTop:
		offsetX = baseBounds.Dx() / 2
		offsetY = 0
	case PositionTopLeft:
		offsetX = 0
		offsetY = 0
	case PositionTopRight:
		offsetX = baseBounds.Dx() - bounds.Dx()
		offsetY = 0
	case PositionBottom:
		offsetX = baseBounds.Dx() / 2
		offsetY = baseBounds.Dy() - bounds.Dy()
	case PositionBottomLeft:
		offsetX = 0
		offsetY = baseBounds.Dy() - bounds.Dy()
	case PositionBottomRight:
		offsetX = baseBounds.Dx()
		offsetY = baseBounds.Dy() - bounds.Dy()
	case PositionCenter:
		offsetX = (baseBounds.Dx() - bounds.Dx()) / 2
		offsetY = (baseBounds.Dy() - bounds.Dy()) / 2
	case PositionLeft:
		offsetX = 0
		offsetY = (baseBounds.Dy() - bounds.Dy()) / 2
	case PositionRight:
		offsetX = baseBounds.Dx() - bounds.Dx()
		offsetY = (baseBounds.Dy() - bounds.Dy()) / 2
	default:
		return fmt.Errorf("invalid position: %s", overlay.Position)
	}
	if overlay.OffsetX != 0 {
		offsetX += int(float64(bounds.Dx()) * overlay.OffsetX)
	}
	if overlay.OffsetY != 0 {
		offsetY += int(float64(bounds.Dy()) * overlay.OffsetY)
	}

	draw.Draw(baseImg, image.Rect(offsetX, offsetY, offsetX+bounds.Dx(), offsetY+bounds.Dy()), img, image.Point{}, draw.Over)

	return nil
}

func resizeImage(baseImg image.Image, img image.Image, scaleX float64, scaleY float64) image.Image {
	bounds := img.Bounds()
	baseBounds := baseImg.Bounds()

	newWidth := bounds.Dx()
	newHeight := bounds.Dy()
	if scaleX != 1 && scaleX != 0 {
		newWidth = int(float64(baseBounds.Dx()) * scaleX)
		// scale the height to keep the aspect ratio
		newHeight = int(float64(newWidth) * float64(bounds.Dy()) / float64(bounds.Dx()))
	} else if scaleY != 1 && scaleY != 0 {
		newHeight = int(float64(baseBounds.Dy()) * scaleY)
		// scale the width to keep the aspect ratio
		newWidth = int(float64(newHeight) * float64(bounds.Dx()) / float64(bounds.Dy()))
	}

	resizedImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.BiLinear.Scale(resizedImg, resizedImg.Bounds(), img, bounds, draw.Src, nil)
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

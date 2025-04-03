package pogoicon

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"

	"golang.org/x/image/draw"
)

type Template struct {
	// Background is the base image, the background sets the size of the final image.
	Background io.Reader `json:"background"`

	Overlays []OverlayTemplate `json:"overlays"`
}

type Position string

const (
	PositionTop      Position = "top"
	PositionTopLeft  Position = "top-left"
	PositionTopRight Position = "top-right"

	PositionBottom      Position = "bottom"
	PositionBottomLeft  Position = "bottom-left"
	PositionBottomRight Position = "bottom-right"

	PositionCenter Position = "center"
	PositionLeft   Position = "left"
	PositionRight  Position = "right"
)

type OverlayTemplate struct {
	// Image is the overlay image.
	Image io.Reader
	// ScaleX is the scale of the overlay image relative to the background image in the horizontal direction.
	// Use 0.0 to keep the original aspect ratio.
	ScaleX float64
	// ScaleY is the scale of the overlay image relative to the background image in the vertical direction.
	// Use 0.0 to keep the original aspect ratio.
	ScaleY float64
	// Position is the position of the overlay image.
	Position Position
	// OffsetX is the x offset of the overlay image.
	OffsetX int
	// OffsetY is the y offset of the overlay image.
	OffsetY int
	// Opacity is the opacity of the overlay image.
	// 0.0 is fully transparent, 1.0 is fully opaque.
	Opacity float64
	// FlipX is whether to flip the overlay image horizontally.
	FlipX bool
	// FlipY is whether to flip the overlay image vertically.
	FlipY bool
	// Rotate is the rotation of the overlay image in degrees.
	Rotate float64
}

func Generate(background io.Reader, backgroundIcon io.Reader, backgroundIconScale float64, pokemon io.Reader, pokemonScale float64, cosmetic io.Reader, cosmeticScale float64) (io.Reader, error) {
	backgroundImage, _, err := image.Decode(background)
	if err != nil {
		return nil, fmt.Errorf("failed to decode background image: %w", err)
	}

	var backgroundIconImage image.Image
	if backgroundIcon != nil {
		backgroundIconImage, _, err = image.Decode(backgroundIcon)
		if err != nil {
			return nil, fmt.Errorf("failed to decode background icon image: %w", err)
		}
	}

	pokemonImage, _, err := image.Decode(pokemon)
	if err != nil {
		return nil, fmt.Errorf("failed to decode pokemon image: %w", err)
	}

	var cosmeticImage image.Image
	if cosmetic != nil {
		cosmeticImage, _, err = image.Decode(cosmetic)
		if err != nil {
			return nil, fmt.Errorf("failed to decode cosmetic image: %w", err)
		}
	}

	newImage := image.NewRGBA(backgroundImage.Bounds())
	draw.Draw(newImage, newImage.Bounds(), backgroundImage, image.Point{}, draw.Src)

	if backgroundIconImage != nil {
		scaledBackgroundIconHeight := int(float64(backgroundImage.Bounds().Dy()) * backgroundIconScale)
		scaledBackgroundIconWidth := int(float64(backgroundIconImage.Bounds().Dx()) * (float64(scaledBackgroundIconHeight) / float64(backgroundIconImage.Bounds().Dy())))

		scaledBackgroundIcon := image.NewRGBA(image.Rect(0, 0, scaledBackgroundIconWidth, scaledBackgroundIconHeight))
		draw.BiLinear.Scale(scaledBackgroundIcon, scaledBackgroundIcon.Bounds(), backgroundIconImage, backgroundIconImage.Bounds(), draw.Over, nil)

		backgroundIconXOffset := (backgroundImage.Bounds().Dx() - scaledBackgroundIconWidth) / 2
		backgroundIconYOffset := (backgroundImage.Bounds().Dy() - scaledBackgroundIconHeight) / 2
		draw.Draw(newImage, image.Rect(backgroundIconXOffset, backgroundIconYOffset, backgroundIconXOffset+scaledBackgroundIconWidth, backgroundIconYOffset+scaledBackgroundIconHeight), scaledBackgroundIcon, image.Point{}, draw.Over)
	}

	scaledHeight := int(float64(backgroundImage.Bounds().Dy()) * pokemonScale)
	scaledWidth := int(float64(pokemonImage.Bounds().Dx()) * pokemonScale)

	scaledPokemon := image.NewRGBA(image.Rect(0, 0, scaledWidth, scaledHeight))
	draw.BiLinear.Scale(scaledPokemon, scaledPokemon.Bounds(), pokemonImage, pokemonImage.Bounds(), draw.Over, nil)

	pokemonXOffset := (backgroundImage.Bounds().Dx() - scaledWidth) / 2
	pokemonYOffset := (backgroundImage.Bounds().Dy() - scaledHeight) / 2
	draw.Draw(newImage, image.Rect(pokemonXOffset, pokemonYOffset, pokemonXOffset+scaledWidth, pokemonYOffset+scaledHeight), scaledPokemon, image.Point{}, draw.Over)

	if cosmeticImage != nil {
		cosmeticHeight := int(float64(backgroundImage.Bounds().Dy()) * cosmeticScale)
		cosmeticWidth := int(float64(cosmeticImage.Bounds().Dx()) * (float64(cosmeticHeight) / float64(cosmeticImage.Bounds().Dy())))

		scaledCosmetic := image.NewRGBA(image.Rect(0, 0, cosmeticWidth, cosmeticHeight))
		draw.BiLinear.Scale(scaledCosmetic, scaledCosmetic.Bounds(), cosmeticImage, cosmeticImage.Bounds(), draw.Over, nil)

		cosmeticXOffset := pokemonXOffset
		cosmeticYOffset := pokemonYOffset
		draw.Draw(newImage, image.Rect(cosmeticXOffset, cosmeticYOffset, cosmeticXOffset+cosmeticWidth, cosmeticYOffset+cosmeticHeight), scaledCosmetic, image.Point{}, draw.Over)
	}

	buf := new(bytes.Buffer)
	if err = png.Encode(buf, newImage); err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}
	return bytes.NewReader(buf.Bytes()), nil
}

func overlayTemplate(baseImg *image.RGBA, overlay OverlayTemplate) error {
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
}

func resizeImage(img image.Image, scaleX float64, scaleY float64) image.Image {
	bounds := img.Bounds()
	newWidth := int(float64(bounds.Dx()) * scaleX)
	newHeight := int(float64(bounds.Dy()) * scaleY)

	resizedImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.BiLinear.Scale(resizedImg, resizedImg.Bounds(), img, bounds, draw.Over, nil)
	return resizedImg
}

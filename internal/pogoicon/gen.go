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
		scaledBackgroundIconWidth := int(float64(backgroundIconImage.Bounds().Dx()) * backgroundIconScale)
		scaledBackgroundIconHeight := int(float64(backgroundIconImage.Bounds().Dy()) * backgroundIconScale)

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
		cosmeticWidth := int(float64(cosmeticImage.Bounds().Dx()) * cosmeticScale)
		cosmeticHeight := int(float64(cosmeticImage.Bounds().Dy()) * cosmeticScale)

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

package icongen

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"math"
	"slices"

	"golang.org/x/image/draw"
)

func Generate(assets fs.FS, cfg Config, pokemonImage func(p string) (io.ReadCloser, error), event string, pokemon []string, cosmetics []string) (io.Reader, error) {
	var eventCfg EventConfig
	for _, e := range cfg.Events {
		if e.Name == event {
			eventCfg = e
			break
		}
	}
	if eventCfg.Name == "" {
		return nil, fmt.Errorf("event %q not found", event)
	}

	layers := eventCfg.Layers

	slices.SortFunc(layers, func(a Layer, b Layer) int {
		if a.ID.Order() == b.ID.Order() {
			return 0
		}
		if a.ID.Order() < b.ID.Order() {
			return -1
		}
		return 1
	})

	index := slices.IndexFunc(layers, func(o Layer) bool {
		return o.ID == LayerIDCosmetic
	})
	if index == -1 {
		index = len(layers)
	}

	pokemonLayers := make([]imageLayer, 0, len(pokemon))
	if len(pokemon) > 0 {
		pLayers := cfg.PokemonLayers[len(pokemon)-1].Layers
		for i, p := range pokemon {
			img, err := pokemonImage(p)
			if err != nil {
				return nil, fmt.Errorf("failed to get pokemon image: %w", err)
			}
			defer img.Close()
			pLayer := pLayers[i]
			pLayer.Image = p
			pokemonLayers = append(pokemonLayers, imageLayer{
				Image: img,
				Layer: pLayer,
			})
		}
	}

	imgLayers := make([]imageLayer, 0, len(layers))
	for _, layer := range layers {
		img, err := assets.Open(layer.Image)
		if err != nil {
			return nil, fmt.Errorf("failed to open layer image: %w", err)
		}
		defer img.Close()

		imgLayers = append(imgLayers, imageLayer{
			Image: img,
			Layer: layer,
		})
	}

	imgLayers = slices.Insert(imgLayers, index, pokemonLayers...)

	for _, c := range cosmetics {
		i := slices.IndexFunc(cfg.Cosmetics, func(config CosmeticConfig) bool {
			return config.Name == c
		})
		if i == -1 {
			return nil, fmt.Errorf("cosmetic %q not found", c)
		}

		for _, layer := range cfg.Cosmetics[i].Layers {
			img, err := assets.Open(layer.Image)
			if err != nil {
				return nil, fmt.Errorf("failed to open cosmetic image: %w", err)
			}
			defer img.Close()

			imgLayers = append(imgLayers, imageLayer{
				Image: img,
				Layer: layer,
			})
		}
	}

	var newImage *image.RGBA
	for i, layer := range imgLayers {
		img, _, err := image.Decode(layer.Image)
		if err != nil {
			return nil, fmt.Errorf("failed to decode image %q: %w", layer.Layer.Image, err)
		}
		if i == 0 {
			newImage = image.NewRGBA(img.Bounds())
		}

		if err = applyOverlay(newImage, img, layer); err != nil {
			return nil, fmt.Errorf("failed to layer template: %w", err)
		}
	}

	buf := new(bytes.Buffer)
	if err := png.Encode(buf, newImage); err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}
	return bytes.NewReader(buf.Bytes()), nil
}

func applyOverlay(baseImg *image.RGBA, img image.Image, layer imageLayer) error {
	img = resizeLayer(baseImg, img, layer.ScaleX, layer.ScaleY)
	img = flipLayer(img, layer.FlipX, layer.FlipY)
	img = rotateLayer(img, layer.Rotate)

	bounds := img.Bounds()
	baseBounds := baseImg.Bounds()
	var (
		offsetX int
		offsetY int
	)
	switch layer.Position {
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
		return fmt.Errorf("invalid layer position: %s", layer.Position)
	}
	if layer.OffsetX != 0 {
		offsetX += int(float64(bounds.Dx()) * layer.OffsetX)
	}
	if layer.OffsetY != 0 {
		offsetY += int(float64(bounds.Dy()) * layer.OffsetY)
	}

	draw.Draw(baseImg, image.Rect(offsetX, offsetY, offsetX+bounds.Dx(), offsetY+bounds.Dy()), img, image.Point{}, draw.Over)

	return nil
}

func resizeLayer(baseImg image.Image, img image.Image, scaleX float64, scaleY float64) image.Image {
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

func flipLayer(img image.Image, flipX bool, flipY bool) image.Image {
	if flipX {
		img = flipXLayer(img)
	}
	if flipY {
		img = flipYLayer(img)
	}
	return img
}

func flipXLayer(img image.Image) image.Image {
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			newImg.Set(bounds.Dx()-1-x, y, img.At(x, y))
		}
	}
	return newImg
}

func flipYLayer(img image.Image) image.Image {
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			newImg.Set(x, bounds.Dy()-1-y, img.At(x, y))
		}
	}
	return newImg
}

func rotateLayer(img image.Image, angle float64) image.Image {
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)

	angle = angle * (math.Pi / 180.0) // Convert degrees to radians
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			newX := int(float64(x)*math.Cos(angle) - float64(y)*math.Sin(angle))
			newY := int(float64(x)*math.Sin(angle) + float64(y)*math.Cos(angle))
			newImg.Set(newX, newY, img.At(x, y))
		}
	}
	return newImg
}

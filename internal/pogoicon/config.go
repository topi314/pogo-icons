package pogoicon

import (
	"io"
)

type Config struct {
	Events        []EventConfig    `toml:"events"`
	Cosmetics     []CosmeticConfig `toml:"cosmetics"`
	PokemonLayers []PokemonConfig  `toml:"pokemon_layers"`
}

type EventConfig struct {
	Name   string  `toml:"name"`
	Layers []Layer `toml:"layers"`
}

type CosmeticConfig struct {
	Name   string  `toml:"name"`
	Layers []Layer `toml:"layers"`
}

type PokemonConfig struct {
	Layers []Layer `toml:"layers"`
}

type LayerID string

const (
	LayerIDBackground LayerID = "background"
	LayerIDPokemon    LayerID = "pokemon"
	LayerIDCosmetic   LayerID = "cosmetic"
)

func (l LayerID) Order() int {
	switch l {
	case LayerIDBackground:
		return 0
	case LayerIDPokemon:
		return 1
	case LayerIDCosmetic:
		return 2
	default:
		return -1
	}
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

// Layer represents an overlay image to be applied to the background image.
type Layer struct {
	// ID is the ID of the overlay.
	ID LayerID `toml:"id"`
	// Image is the asset path of the overlay image.
	Image string `toml:"image"`
	// ScaleX is the scale of the overlay image relative to the background image in the horizontal direction.
	// Use 0.0 to keep the original aspect ratio.
	ScaleX float64 `toml:"scale_x"`
	// ScaleY is the scale of the overlay image relative to the background image in the vertical direction.
	// Use 0.0 to keep the original aspect ratio.
	ScaleY float64 `toml:"scale_y"`
	// Position is the position of the overlay image.
	Position Position `toml:"position"`
	// OffsetX is the x offset of the overlay image.
	OffsetX float64 `toml:"offset_x"`
	// OffsetY is the y offset of the overlay image.
	OffsetY float64 `toml:"offset_y"`
	// FlipX is whether to flip the overlay image horizontally.
	FlipX bool `toml:"flip_x"`
	// FlipY is whether to flip the overlay image vertically.
	FlipY bool `toml:"flip_y"`
	// Rotate is the rotation of the overlay image in degrees.
	Rotate float64 `toml:"rotate"`
}

type imageLayer struct {
	Image io.Reader
	Layer
}

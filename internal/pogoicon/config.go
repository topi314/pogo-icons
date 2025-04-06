package pogoicon

import (
	"io"
)

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

type Overlay struct {
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
	// FlipX is whether to flip the overlay image horizontally.
	FlipX bool
	// FlipY is whether to flip the overlay image vertically.
	FlipY bool
	// Rotate is the rotation of the overlay image in degrees.
	Rotate float64
}

type AssetConfig struct {
	Events    []EventConfig    `toml:"events"`
	Cosmetics []CosmeticConfig `toml:"cosmetics"`
}

type EventConfig struct {
	Name       string          `toml:"name"`
	Background string          `toml:"background"`
	Overlays   []OverlayConfig `toml:"overlays"`
}

type OverlayConfig struct {
	Image    string
	ScaleX   float64
	ScaleY   float64
	Position Position
	OffsetX  int
	OffsetY  int
	FlipX    bool
	FlipY    bool
	Rotate   float64
}

type CosmeticConfig struct {
	Name  string  `toml:"name"`
	Image string  `toml:"image"`
	Scale float64 `toml:"scale"`
}

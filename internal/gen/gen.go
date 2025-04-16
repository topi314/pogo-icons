package gen

import (
	"fmt"
	"io"
	"io/fs"
	"slices"

	"github.com/topi314/pogo-icons/internal/pogoicon"
)

var pokemonOverlayConfigs = []func() []pogoicon.Overlay{
	func() []pogoicon.Overlay {
		return []pogoicon.Overlay{
			{
				Position: pogoicon.PositionCenter,
			},
		}
	},
	func() []pogoicon.Overlay {
		return []pogoicon.Overlay{
			{
				Position: pogoicon.PositionCenter,
				ScaleY:   0.6,
				OffsetX:  -0.4,
			},
			{
				Position: pogoicon.PositionCenter,
				ScaleY:   0.6,
				OffsetX:  0.4,
			},
		}
	},
	func() []pogoicon.Overlay {
		return []pogoicon.Overlay{
			{
				Position: pogoicon.PositionCenter,
				ScaleY:   0.6,
				OffsetY:  -0.4,
			},
			{
				Position: pogoicon.PositionCenter,
				ScaleY:   0.6,
				OffsetX:  0.4,
				OffsetY:  0.4,
			},
			{
				Position: pogoicon.PositionCenter,
				ScaleY:   0.6,
				OffsetX:  -0.4,
				OffsetY:  0.4,
			},
		}
	},
	func() []pogoicon.Overlay {
		return []pogoicon.Overlay{
			{
				Position: pogoicon.PositionCenter,
				ScaleY:   0.6,
				OffsetX:  -0.4,
				OffsetY:  -0.4,
			},
			{
				Position: pogoicon.PositionCenter,
				ScaleY:   0.6,
				OffsetX:  0.4,
				OffsetY:  -0.4,
			},
			{
				Position: pogoicon.PositionCenter,
				ScaleY:   0.6,
				OffsetX:  0.4,
				OffsetY:  0.4,
			},
			{
				Position: pogoicon.PositionCenter,
				ScaleY:   0.6,
				OffsetX:  -0.4,
				OffsetY:  0.4,
			},
		}
	},
}

type LayerID int

const (
	LayerIDBackground LayerID = 0
	LayerIDPokemon    LayerID = 1
	LayerIDCosmetic   LayerID = 2
)

type OverlayConfig struct {
	ID       LayerID
	Image    string
	ScaleX   float64
	ScaleY   float64
	Position pogoicon.Position
	OffsetX  float64
	OffsetY  float64
	FlipX    bool
	FlipY    bool
	Rotate   float64
}

func Generate(assets fs.FS, pokemonImage func(pokemon string) (io.ReadCloser, error), pokemon []string, overlays []OverlayConfig) (io.Reader, error) {
	slices.SortFunc(overlays, func(a OverlayConfig, b OverlayConfig) int {
		if a.ID == b.ID {
			return 0
		}
		if a.ID < b.ID {
			return -1
		}
		return 1
	})

	index := slices.IndexFunc(overlays, func(o OverlayConfig) bool {
		return o.ID == LayerIDCosmetic
	})
	if index == -1 {
		index = len(overlays)
	}

	pokemonOverlays := make([]pogoicon.Overlay, 0, len(pokemon))
	overlayConfigs := pokemonOverlayConfigs[len(pokemon)-1]()
	for i, p := range pokemon {
		img, err := pokemonImage(p)
		if err != nil {
			return nil, fmt.Errorf("failed to get pokemon image: %w", err)
		}
		defer img.Close()

		overlayConfigs[i].Image = img

		pokemonOverlays = append(pokemonOverlays, overlayConfigs[i])
	}

	pogoOverlays := make([]pogoicon.Overlay, 0, len(overlays))
	for _, overlay := range overlays {
		if overlay.ID == LayerIDPokemon {
			continue
		}

		img, err := assets.Open(overlay.Image)
		if err != nil {
			return nil, fmt.Errorf("failed to open overlay image: %w", err)
		}
		defer img.Close()

		pogoOverlays = append(pogoOverlays, pogoicon.Overlay{
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

	pogoOverlays = slices.Insert(pogoOverlays, index, pokemonOverlays...)

	return pogoicon.Generate(pogoOverlays)
}

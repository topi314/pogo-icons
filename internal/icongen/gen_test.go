package icongen

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/topi314/pogo-icons/internal/pokeapi"
)

func TestGenerate(t *testing.T) {
	client := pokeapi.NewAPI("https://pokeapi.co/api/v2/")

	event := "test"
	pokemon := []string{"venusaur", "charizard", "blastoise"}
	cosmetics := []string{"CA Star"}
	cfg := Config{
		Events: []EventConfig{
			{
				Name: "test",
				Layers: []Layer{
					{
						ID:       LayerIDBackground,
						Image:    "backgrounds/generic_day.png",
						Position: PositionTopLeft,
					},
					{
						ID:       LayerIDCosmetic,
						Image:    "icons/ca_star.png",
						Position: PositionTopLeft,
						ScaleY:   0.2,
						OffsetX:  2.5,
					},
				},
			},
		},
		Cosmetics: []CosmeticConfig{
			{
				Name: "CA Star",
				Layers: []Layer{
					{
						ID:       LayerIDCosmetic,
						Image:    "icons/ca_star.png",
						Position: PositionTopLeft,
						ScaleY:   0.2,
						OffsetX:  2.5,
					},
				},
			},
		},
		PokemonLayers: []PokemonConfig{
			{
				Layers: []Layer{
					{
						ID:       LayerIDPokemon,
						Position: PositionCenter,
					},
				},
			},
			{
				Layers: []Layer{
					{
						ID:       LayerIDPokemon,
						Position: PositionCenter,
						ScaleY:   0.6,
						OffsetX:  -0.4,
					},
					{
						ID:       LayerIDPokemon,
						Position: PositionCenter,
						ScaleY:   0.6,
						OffsetX:  0.4,
					},
				},
			},
			{
				Layers: []Layer{
					{
						ID:       LayerIDPokemon,
						Position: PositionCenter,
						ScaleY:   0.6,
						OffsetY:  -0.4,
					},
					{
						ID:       LayerIDPokemon,
						Position: PositionCenter,
						ScaleY:   0.6,
						OffsetX:  0.4,
						OffsetY:  0.4,
					},
					{
						ID:       LayerIDPokemon,
						Position: PositionCenter,
						ScaleY:   0.6,
						OffsetX:  -0.4,
						OffsetY:  0.4,
					},
				},
			},
		},
	}

	var getPokemonImage = func(p string) (io.ReadCloser, error) {
		pf, err := client.GetPokemonForm(context.Background(), p)
		if err != nil {
			return nil, err
		}
		sprite, err := client.GetSprite(context.Background(), pf.Sprite)
		if err != nil {
			return nil, err
		}
		return sprite.Body, nil
	}

	assets := os.DirFS("../../assets")

	img, err := Generate(assets, cfg, getPokemonImage, event, pokemon, cosmetics)
	if err != nil {
		t.Fatalf("failed to generate image: %v", err)
	}

	output, err := os.OpenFile("output.png", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		t.Fatalf("failed to open output file: %v", err)
	}
	defer output.Close()

	_, err = io.Copy(output, img)
	if err != nil {
		t.Fatalf("failed to write output file: %v", err)
	}
	t.Logf("Image generated successfully and saved to output.png")
}

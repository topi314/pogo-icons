package gen

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/topi314/pogo-icons/internal/pogoicon"
	"github.com/topi314/pogo-icons/internal/pokeapi"
)

func TestGenerate(t *testing.T) {
	client := pokeapi.NewAPI("https://pokeapi.co/api/v2/")

	var getPokemonImage = func(pokemon string) (io.ReadCloser, error) {
		p, err := client.GetPokemonForm(context.Background(), pokemon)
		if err != nil {
			return nil, err
		}
		t.Logf("sprite: %s", p.Sprite)
		sprite, err := client.GetSprite(context.Background(), p.Sprite)
		if err != nil {
			return nil, err
		}
		return sprite.Body, nil
	}

	pokemon := []string{"charizard", "blastoise"}
	overlays := []OverlayConfig{
		{
			ID:       LayerIDBackground,
			Image:    "backgrounds/generic_day.png",
			Position: pogoicon.PositionTopLeft,
		},
		{
			ID:       LayerIDCosmetic,
			Image:    "icons/ca_star.png",
			Position: pogoicon.PositionTopLeft,
			ScaleY:   0.2,
			OffsetX:  2.5,
		},
	}

	assets := os.DirFS("../../assets")

	img, err := Generate(assets, getPokemonImage, pokemon, overlays)
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

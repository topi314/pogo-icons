package pokeapi

import (
	"strings"
)

func newPokemonForm(p Pokemon) PokemonForm {
	return PokemonForm{
		Name:        strings.Title(strings.ReplaceAll(p.Name, "-", " ")),
		Value:       p.Name,
		Sprite:      p.Sprites.Other.OfficialArtwork.FrontDefault,
		ShinySprite: p.Sprites.Other.OfficialArtwork.FrontShiny,
	}
}

type PokemonForm struct {
	Name        string
	Value       string
	Sprite      string
	ShinySprite string
}

func (f PokemonForm) FilterValue() string {
	return f.Name
}

type Page[T any] struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []T    `json:"results"`
}

type PokemonSpecie struct {
	BaseHappiness int `json:"base_happiness"`
	CaptureRate   int `json:"capture_rate"`
	Color         struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"color"`
	EggGroups []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"egg_groups"`
	EvolutionChain struct {
		URL string `json:"url"`
	} `json:"evolution_chain"`
	EvolvesFromSpecies struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"evolves_from_species"`
	FlavorTextEntries []struct {
		FlavorText string `json:"flavor_text"`
		Language   struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"language"`
		Version struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"version"`
	} `json:"flavor_text_entries"`
	FormDescriptions []interface{} `json:"form_descriptions"`
	FormsSwitchable  bool          `json:"forms_switchable"`
	GenderRate       int           `json:"gender_rate"`
	Genera           []struct {
		Genus    string `json:"genus"`
		Language struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"language"`
	} `json:"genera"`
	Generation struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"generation"`
	GrowthRate struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"growth_rate"`
	Habitat struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"habitat"`
	HasGenderDifferences bool   `json:"has_gender_differences"`
	HatchCounter         int    `json:"hatch_counter"`
	ID                   int    `json:"id"`
	IsBaby               bool   `json:"is_baby"`
	IsLegendary          bool   `json:"is_legendary"`
	IsMythical           bool   `json:"is_mythical"`
	Name                 string `json:"name"`
	Names                []struct {
		Language struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"language"`
		Name string `json:"name"`
	} `json:"names"`
	Order             int `json:"order"`
	PalParkEncounters []struct {
		Area struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"area"`
		BaseScore int `json:"base_score"`
		Rate      int `json:"rate"`
	} `json:"pal_park_encounters"`
	PokedexNumbers []struct {
		EntryNumber int `json:"entry_number"`
		Pokedex     struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokedex"`
	} `json:"pokedex_numbers"`
	Shape struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"shape"`
	Varieties []struct {
		IsDefault bool `json:"is_default"`
		Pokemon   struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
	} `json:"varieties"`
}

type Pokemon struct {
	Abilities []struct {
		Ability struct {
			Name string `json:"name"`
			Url  string `json:"url"`
		} `json:"ability"`
		IsHidden bool `json:"is_hidden"`
		Slot     int  `json:"slot"`
	} `json:"abilities"`
	BaseExperience int `json:"base_experience"`
	Cries          struct {
		Latest string      `json:"latest"`
		Legacy interface{} `json:"legacy"`
	} `json:"cries"`
	Forms []struct {
		Name string `json:"name"`
		Url  string `json:"url"`
	} `json:"forms"`
	GameIndices            []interface{} `json:"game_indices"`
	Height                 int           `json:"height"`
	HeldItems              []interface{} `json:"held_items"`
	ID                     int           `json:"id"`
	IsDefault              bool          `json:"is_default"`
	LocationAreaEncounters string        `json:"location_area_encounters"`
	Moves                  []interface{} `json:"moves"`
	Name                   string        `json:"name"`
	Order                  int           `json:"order"`
	PastAbilities          []interface{} `json:"past_abilities"`
	PastTypes              []interface{} `json:"past_types"`
	Species                struct {
		Name string `json:"name"`
		Url  string `json:"url"`
	} `json:"species"`
	Sprites struct {
		BackDefault      string      `json:"back_default"`
		BackFemale       interface{} `json:"back_female"`
		BackShiny        string      `json:"back_shiny"`
		BackShinyFemale  interface{} `json:"back_shiny_female"`
		FrontDefault     string      `json:"front_default"`
		FrontFemale      interface{} `json:"front_female"`
		FrontShiny       string      `json:"front_shiny"`
		FrontShinyFemale interface{} `json:"front_shiny_female"`
		Other            struct {
			DreamWorld struct {
				FrontDefault interface{} `json:"front_default"`
				FrontFemale  interface{} `json:"front_female"`
			} `json:"dream_world"`
			Home struct {
				FrontDefault     string      `json:"front_default"`
				FrontFemale      interface{} `json:"front_female"`
				FrontShiny       string      `json:"front_shiny"`
				FrontShinyFemale interface{} `json:"front_shiny_female"`
			} `json:"home"`
			OfficialArtwork struct {
				FrontDefault string `json:"front_default"`
				FrontShiny   string `json:"front_shiny"`
			} `json:"official-artwork"`
			Showdown struct {
				BackDefault      interface{} `json:"back_default"`
				BackFemale       interface{} `json:"back_female"`
				BackShiny        interface{} `json:"back_shiny"`
				BackShinyFemale  interface{} `json:"back_shiny_female"`
				FrontDefault     string      `json:"front_default"`
				FrontFemale      interface{} `json:"front_female"`
				FrontShiny       string      `json:"front_shiny"`
				FrontShinyFemale interface{} `json:"front_shiny_female"`
			} `json:"showdown"`
		} `json:"other"`
		Versions struct {
			GenerationI struct {
				RedBlue struct {
					BackDefault      interface{} `json:"back_default"`
					BackGray         interface{} `json:"back_gray"`
					BackTransparent  interface{} `json:"back_transparent"`
					FrontDefault     interface{} `json:"front_default"`
					FrontGray        interface{} `json:"front_gray"`
					FrontTransparent interface{} `json:"front_transparent"`
				} `json:"red-blue"`
				Yellow struct {
					BackDefault      interface{} `json:"back_default"`
					BackGray         interface{} `json:"back_gray"`
					BackTransparent  interface{} `json:"back_transparent"`
					FrontDefault     interface{} `json:"front_default"`
					FrontGray        interface{} `json:"front_gray"`
					FrontTransparent interface{} `json:"front_transparent"`
				} `json:"yellow"`
			} `json:"generation-i"`
			GenerationIi struct {
				Crystal struct {
					BackDefault           interface{} `json:"back_default"`
					BackShiny             interface{} `json:"back_shiny"`
					BackShinyTransparent  interface{} `json:"back_shiny_transparent"`
					BackTransparent       interface{} `json:"back_transparent"`
					FrontDefault          interface{} `json:"front_default"`
					FrontShiny            interface{} `json:"front_shiny"`
					FrontShinyTransparent interface{} `json:"front_shiny_transparent"`
					FrontTransparent      interface{} `json:"front_transparent"`
				} `json:"crystal"`
				Gold struct {
					BackDefault      interface{} `json:"back_default"`
					BackShiny        interface{} `json:"back_shiny"`
					FrontDefault     interface{} `json:"front_default"`
					FrontShiny       interface{} `json:"front_shiny"`
					FrontTransparent interface{} `json:"front_transparent"`
				} `json:"gold"`
				Silver struct {
					BackDefault      interface{} `json:"back_default"`
					BackShiny        interface{} `json:"back_shiny"`
					FrontDefault     interface{} `json:"front_default"`
					FrontShiny       interface{} `json:"front_shiny"`
					FrontTransparent interface{} `json:"front_transparent"`
				} `json:"silver"`
			} `json:"generation-ii"`
			GenerationIii struct {
				Emerald struct {
					FrontDefault interface{} `json:"front_default"`
					FrontShiny   interface{} `json:"front_shiny"`
				} `json:"emerald"`
				FireredLeafgreen struct {
					BackDefault  interface{} `json:"back_default"`
					BackShiny    interface{} `json:"back_shiny"`
					FrontDefault interface{} `json:"front_default"`
					FrontShiny   interface{} `json:"front_shiny"`
				} `json:"firered-leafgreen"`
				RubySapphire struct {
					BackDefault  interface{} `json:"back_default"`
					BackShiny    interface{} `json:"back_shiny"`
					FrontDefault interface{} `json:"front_default"`
					FrontShiny   interface{} `json:"front_shiny"`
				} `json:"ruby-sapphire"`
			} `json:"generation-iii"`
			GenerationIv struct {
				DiamondPearl struct {
					BackDefault      interface{} `json:"back_default"`
					BackFemale       interface{} `json:"back_female"`
					BackShiny        interface{} `json:"back_shiny"`
					BackShinyFemale  interface{} `json:"back_shiny_female"`
					FrontDefault     interface{} `json:"front_default"`
					FrontFemale      interface{} `json:"front_female"`
					FrontShiny       interface{} `json:"front_shiny"`
					FrontShinyFemale interface{} `json:"front_shiny_female"`
				} `json:"diamond-pearl"`
				HeartgoldSoulsilver struct {
					BackDefault      interface{} `json:"back_default"`
					BackFemale       interface{} `json:"back_female"`
					BackShiny        interface{} `json:"back_shiny"`
					BackShinyFemale  interface{} `json:"back_shiny_female"`
					FrontDefault     interface{} `json:"front_default"`
					FrontFemale      interface{} `json:"front_female"`
					FrontShiny       interface{} `json:"front_shiny"`
					FrontShinyFemale interface{} `json:"front_shiny_female"`
				} `json:"heartgold-soulsilver"`
				Platinum struct {
					BackDefault      interface{} `json:"back_default"`
					BackFemale       interface{} `json:"back_female"`
					BackShiny        interface{} `json:"back_shiny"`
					BackShinyFemale  interface{} `json:"back_shiny_female"`
					FrontDefault     interface{} `json:"front_default"`
					FrontFemale      interface{} `json:"front_female"`
					FrontShiny       interface{} `json:"front_shiny"`
					FrontShinyFemale interface{} `json:"front_shiny_female"`
				} `json:"platinum"`
			} `json:"generation-iv"`
			GenerationV struct {
				BlackWhite struct {
					Animated struct {
						BackDefault      interface{} `json:"back_default"`
						BackFemale       interface{} `json:"back_female"`
						BackShiny        interface{} `json:"back_shiny"`
						BackShinyFemale  interface{} `json:"back_shiny_female"`
						FrontDefault     interface{} `json:"front_default"`
						FrontFemale      interface{} `json:"front_female"`
						FrontShiny       interface{} `json:"front_shiny"`
						FrontShinyFemale interface{} `json:"front_shiny_female"`
					} `json:"animated"`
					BackDefault      string      `json:"back_default"`
					BackFemale       interface{} `json:"back_female"`
					BackShiny        string      `json:"back_shiny"`
					BackShinyFemale  interface{} `json:"back_shiny_female"`
					FrontDefault     string      `json:"front_default"`
					FrontFemale      interface{} `json:"front_female"`
					FrontShiny       string      `json:"front_shiny"`
					FrontShinyFemale interface{} `json:"front_shiny_female"`
				} `json:"black-white"`
			} `json:"generation-v"`
			GenerationVi struct {
				OmegarubyAlphasapphire struct {
					FrontDefault     interface{} `json:"front_default"`
					FrontFemale      interface{} `json:"front_female"`
					FrontShiny       interface{} `json:"front_shiny"`
					FrontShinyFemale interface{} `json:"front_shiny_female"`
				} `json:"omegaruby-alphasapphire"`
				XY struct {
					FrontDefault     interface{} `json:"front_default"`
					FrontFemale      interface{} `json:"front_female"`
					FrontShiny       interface{} `json:"front_shiny"`
					FrontShinyFemale interface{} `json:"front_shiny_female"`
				} `json:"x-y"`
			} `json:"generation-vi"`
			GenerationVii struct {
				Icons struct {
					FrontDefault interface{} `json:"front_default"`
					FrontFemale  interface{} `json:"front_female"`
				} `json:"icons"`
				UltraSunUltraMoon struct {
					FrontDefault     interface{} `json:"front_default"`
					FrontFemale      interface{} `json:"front_female"`
					FrontShiny       interface{} `json:"front_shiny"`
					FrontShinyFemale interface{} `json:"front_shiny_female"`
				} `json:"ultra-sun-ultra-moon"`
			} `json:"generation-vii"`
			GenerationViii struct {
				Icons struct {
					FrontDefault string      `json:"front_default"`
					FrontFemale  interface{} `json:"front_female"`
				} `json:"icons"`
			} `json:"generation-viii"`
		} `json:"versions"`
	} `json:"sprites"`
	Stats []struct {
		BaseStat int `json:"base_stat"`
		Effort   int `json:"effort"`
		Stat     struct {
			Name string `json:"name"`
			Url  string `json:"url"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Slot int `json:"slot"`
		Type struct {
			Name string `json:"name"`
			Url  string `json:"url"`
		} `json:"type"`
	} `json:"types"`
	Weight int `json:"weight"`
}

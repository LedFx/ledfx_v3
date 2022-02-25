package api

import (
	"encoding/json"
	"fmt"
	"ledfx/logger"
	"net/http"
)

func HandleColors() {
	http.HandleFunc("/api/colors", func(w http.ResponseWriter, r *http.Request) {
		SetHeader(w)

		rawIn := json.RawMessage(`
		{
			"colors": {
				"builtin": {
					"red": "#ff0000",
					"orange-deep": "#ff2800",
					"orange": "#ff7800",
					"yellow": "#ffc800",
					"yellow-acid": "#a0ff00",
					"green": "#00ff00",
					"green-forest": "#228b22",
					"green-spring": "#00ff7f",
					"green-teal": "#008080",
					"green-turquoise": "#00c78c",
					"green-coral": "#00ff32",
					"cyan": "#00ffff",
					"blue": "#0000ff",
					"blue-light": "#4169e1",
					"blue-navy": "#000080",
					"blue-aqua": "#00ffff",
					"purple": "#800080",
					"pink": "#ff00b2",
					"magenta": "#ff00ff",
					"black": "#000000",
					"white": "#ffffff",
					"gold": "#ffd700",
					"hotpink": "#ff69b4",
					"lightblue": "#add8e6",
					"lightgreen": "#98fb98",
					"lightpink": "#ffb6c1",
					"lightyellow": "#ffffe0",
					"maroon": "#800000",
					"mint": "#bdfcc9",
					"olive": "#556b2f",
					"peach": "#ff6464",
					"plum": "#dda0dd",
					"sepia": "#5e2612",
					"skyblue": "#87ceeb",
					"steelblue": "#4682b4",
					"tan": "#d2b48c",
					"violetred": "#d02090"
				},
				"user": {}
			},
			"gradients": {
				"builtin": {
					"Rainbow": "linear-gradient(90deg, rgb(255, 0, 0) 0%, rgb(255, 120, 0) 14%, rgb(255, 200, 0) 28%, rgb(0, 255, 0) 42%, rgb(0, 199, 140) 56%, rgb(0, 0, 255) 70%, rgb(128, 0, 128) 84%, rgb(255, 0, 178) 98%)",
					"Dancefloor": "linear-gradient(90deg, rgb(255, 0, 0) 0%, rgb(255, 0, 178) 50%, rgb(0, 0, 255) 100%)",
					"Plasma": "linear-gradient(90deg, rgb(0, 0, 255) 0%, rgb(128, 0, 128) 25%, rgb(255, 0, 0) 50%, rgb(255, 40, 0) 75%, rgb(255, 200, 0) 100%)",
					"Ocean": "linear-gradient(90deg, rgb(0, 255, 255) 0%, rgb(0, 0, 255) 100%)",
					"Viridis": "linear-gradient(90deg, rgb(128, 0, 128) 0%, rgb(0, 0, 255) 25%, rgb(0, 128, 128) 50%, rgb(0, 255, 0) 75%, rgb(255, 200, 0) 100%)",
					"Jungle": "linear-gradient(90deg, rgb(0, 255, 0) 0%, rgb(34, 139, 34) 50%, rgb(255, 120, 0) 100%)",
					"Spring": "linear-gradient(90deg, rgb(255, 0, 178) 0%, rgb(255, 40, 0) 50%, rgb(255, 200, 0) 100%)",
					"Winter": "linear-gradient(90deg, rgb(0, 199, 140) 0%, rgb(0, 255, 50) 100%)",
					"Frost": "linear-gradient(90deg, rgb(0, 0, 255) 0%, rgb(0, 255, 255) 33%, rgb(128, 0, 128) 66%, rgb(255, 0, 178) 99%)",
					"Sunset": "linear-gradient(90deg, rgb(0, 0, 128) 0%, rgb(255, 120, 0) 50%, rgb(255, 0, 0) 100%)",
					"Borealis": "linear-gradient(90deg, rgb(255, 40, 0) 0%, rgb(128, 0, 128) 33%, rgb(0, 199, 140) 66%, rgb(0, 255, 0) 99%)",
					"Rust": "linear-gradient(90deg, rgb(255, 40, 0) 0%, rgb(255, 0, 0) 100%)",
					"Winamp": "linear-gradient(90deg, rgb(0, 255, 0) 0%, rgb(255, 200, 0) 25%, rgb(255, 120, 0) 50%, rgb(255, 40, 0) 75%, rgb(255, 0, 0) 100%)"
				},
				"user": {}
			}
		}`)
		var objmap map[string]*json.RawMessage
		err := json.Unmarshal(rawIn, &objmap)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(objmap)

		err = json.NewEncoder(w).Encode(objmap)
		if err != nil {
			logger.Logger.Warn(err)
		}
		// json.NewEncoder(w).Encode(config.Schema.Effects)
	})

	http.HandleFunc("/api/effects/singleColor/presets", func(w http.ResponseWriter, r *http.Request) {
		SetHeader(w)

		rawIn := json.RawMessage(`
		{
			"status": "success",
			"effect": "singleColor",
			"default_presets": {},
			"custom_presets": {}
		}`)
		// ToDo: uncomment when preset-handling is there
		// rawIn := json.RawMessage(`
		// {
		// 	"status": "success",
		// 	"effect": "singleColor",
		// 	"default_presets": {
		// 		"reset": {
		// 			"config": {},
		// 			"name": "Reset"
		// 		},
		// 		"blue": {
		// 			"config": {
		// 				"background_brightness": 1.0,
		// 				"background_color": "black",
		// 				"blur": 0,
		// 				"brightness": 1,
		// 				"color": "blue",
		// 				"decay": 1,
		// 				"flip": false,
		// 				"mirror": true,
		// 				"modulate": false,
		// 				"modulation_effect": "sine",
		// 				"modulation_speed": 0.5,
		// 				"speed": 5,
		// 				"threshold": 0.7
		// 			},
		// 			"name": "Blue"
		// 		},
		// 		"cyan": {
		// 			"config": {
		// 				"background_brightness": 1.0,
		// 				"background_color": "black",
		// 				"blur": 0,
		// 				"brightness": 1,
		// 				"color": "cyan",
		// 				"decay": 1,
		// 				"flip": false,
		// 				"mirror": true,
		// 				"modulate": false,
		// 				"modulation_effect": "sine",
		// 				"modulation_speed": 0.5,
		// 				"speed": 5,
		// 				"threshold": 0.7
		// 			},
		// 			"name": "Cyan"
		// 		},
		// 		"green": {
		// 			"config": {
		// 				"background_brightness": 1.0,
		// 				"background_color": "black",
		// 				"blur": 0,
		// 				"brightness": 1,
		// 				"color": "green",
		// 				"decay": 1,
		// 				"flip": false,
		// 				"mirror": true,
		// 				"modulate": false,
		// 				"modulation_effect": "sine",
		// 				"modulation_speed": 0.5,
		// 				"speed": 5,
		// 				"threshold": 0.7
		// 			},
		// 			"name": "Green"
		// 		},
		// 		"magenta": {
		// 			"config": {
		// 				"background_brightness": 1.0,
		// 				"background_color": "black",
		// 				"blur": 0,
		// 				"brightness": 1,
		// 				"color": "magenta",
		// 				"decay": 1,
		// 				"flip": false,
		// 				"mirror": true,
		// 				"modulate": false,
		// 				"modulation_effect": "sine",
		// 				"modulation_speed": 0.5,
		// 				"speed": 5,
		// 				"threshold": 0.7
		// 			},
		// 			"name": "Magenta"
		// 		},
		// 		"orange": {
		// 			"config": {
		// 				"background_brightness": 1.0,
		// 				"background_color": "black",
		// 				"blur": 0,
		// 				"brightness": 1,
		// 				"color": "orange-deep",
		// 				"decay": 1,
		// 				"flip": false,
		// 				"mirror": true,
		// 				"modulate": false,
		// 				"modulation_effect": "sine",
		// 				"modulation_speed": 0.5,
		// 				"speed": 5,
		// 				"threshold": 0.7
		// 			},
		// 			"name": "Orange"
		// 		},
		// 		"pink": {
		// 			"config": {
		// 				"background_brightness": 1.0,
		// 				"background_color": "black",
		// 				"blur": 0,
		// 				"brightness": 1,
		// 				"color": "pink",
		// 				"decay": 1,
		// 				"flip": false,
		// 				"mirror": true,
		// 				"modulate": false,
		// 				"modulation_effect": "sine",
		// 				"modulation_speed": 0.5,
		// 				"speed": 5,
		// 				"threshold": 0.7
		// 			},
		// 			"name": "Pink"
		// 		},
		// 		"red": {
		// 			"config": {
		// 				"background_brightness": 1.0,
		// 				"background_color": "black",
		// 				"blur": 0,
		// 				"brightness": 1,
		// 				"color": "red",
		// 				"decay": 1,
		// 				"flip": false,
		// 				"mirror": true,
		// 				"modulate": false,
		// 				"modulation_effect": "sine",
		// 				"modulation_speed": 0.5,
		// 				"speed": 5,
		// 				"threshold": 0.7
		// 			},
		// 			"name": "Red"
		// 		},
		// 		"red-waves": {
		// 			"config": {
		// 				"background_brightness": 1.0,
		// 				"background_color": "black",
		// 				"blur": 6.2,
		// 				"brightness": 1,
		// 				"color": "red",
		// 				"flip": false,
		// 				"mirror": true,
		// 				"modulate": true,
		// 				"modulation_effect": "sine",
		// 				"modulation_speed": 0.76,
		// 				"speed": 0.62
		// 			},
		// 			"name": "Red Waves"
		// 		},
		// 		"steel-pulse": {
		// 			"config": {
		// 				"background_brightness": 1.0,
		// 				"background_color": "black",
		// 				"blur": 6.2,
		// 				"brightness": 1,
		// 				"color": "steelblue",
		// 				"flip": false,
		// 				"mirror": true,
		// 				"modulate": true,
		// 				"modulation_effect": "breath",
		// 				"modulation_speed": 0.75,
		// 				"speed": 0.62
		// 			},
		// 			"name": "Steel Pulse"
		// 		},
		// 		"turquoise-roll": {
		// 			"config": {
		// 				"background_brightness": 1.0,
		// 				"background_color": "black",
		// 				"blur": 6.2,
		// 				"brightness": 1,
		// 				"color": "green-turquoise",
		// 				"flip": false,
		// 				"mirror": false,
		// 				"modulate": true,
		// 				"modulation_effect": "sine",
		// 				"modulation_speed": 0.76,
		// 				"speed": 0.62
		// 			},
		// 			"name": "Turquoise Roll"
		// 		},
		// 		"yellow": {
		// 			"config": {
		// 				"background_brightness": 1.0,
		// 				"background_color": "black",
		// 				"blur": 0,
		// 				"brightness": 1,
		// 				"color": "yellow",
		// 				"decay": 1,
		// 				"flip": false,
		// 				"mirror": true,
		// 				"modulate": false,
		// 				"modulation_effect": "sine",
		// 				"modulation_speed": 0.5,
		// 				"speed": 5,
		// 				"threshold": 0.7
		// 			},
		// 			"name": "Yellow"
		// 		}
		// 	},
		// 	"custom_presets": {}
		// }`)
		var objmap map[string]*json.RawMessage
		err := json.Unmarshal(rawIn, &objmap)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(objmap)

		err = json.NewEncoder(w).Encode(objmap)
		if err != nil {
			logger.Logger.Warn(err)
		}
		// json.NewEncoder(w).Encode(config.Schema.Effects)
	})
}

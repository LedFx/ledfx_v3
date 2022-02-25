package api

import (
	"encoding/json"
	"fmt"
	"ledfx/logger"
	"net/http"
)

func HandleSchema() {
	http.HandleFunc("/api/schema", func(w http.ResponseWriter, r *http.Request) {
		SetHeader(w)

		rawIn := json.RawMessage(`
	{
		"effects": {
			"singleColor": {
				"schema": {
					"properties": {
						"color": {
							"type": "color",
							"gradient": false,
							"title": "Color",
							"description": "Color of strip",
							"default": "#FF0000"
						}
					}
				},
				"id": "singleColor",
				"name": "Single Color",
				"category": "Non-Reactive"
			},
			"audioRandom": {
				"schema": {
					"properties": {
						"color": {
							"type": "color",
							"gradient": false,
							"title": "Color",
							"description": "Color of strip",
							"default": "#FF0000"
						}
					}
				},
				"id": "audioRandom",
				"name": "Audio Random",
				"category": "Reactive"
			}
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
}

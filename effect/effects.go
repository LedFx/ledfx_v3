package effect

import (
	"encoding/json"
	"fmt"
	"ledfx/color"
	"ledfx/utils"
	"log"
	"reflect"
	"strconv"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
)

/*
example usage
pixelgenerator = ledfx.effects.New("energy", config)
ledfx.virtuals.get("virtual_id").SetEffect(pixelgenerator)
*/

/*
NEW EFFECTS MUST BE REGISTERED IN THESE TWO FUNCTIONS =====================
*/

// Generate a map schema for all effects
func Schema() (schema map[string]interface{}, err error) {
	// Copypaste for new effect types
	schema["energy"], err = utils.CreateSchema(reflect.TypeOf((*Energy)(nil)).Elem())
	return schema, err
}

// Creates a new effect and returns its unique id
func New(effect_type string, config interface{}) (effect PixelGenerator, err error) {
	switch effect_type {
	case "energy":
		effect = new(Energy)
	default:
		return effect, fmt.Errorf("%s is not a known effect type", effect_type)
	}

	// create an id and add it to the internal list of instances
	id := effect_type
	for i := 0; ; i++ {
		id = effect_type + strconv.Itoa(i)
		_, exists := effectInstances[id]
		if !exists {
			effectInstances[id] = effect
			break
		}
	}
	// initialise the new effect with its id and config
	effect.Initialize()
	effect.UpdateConfig(config)
	return effect, nil
}

/*
Nothing to modify below here =====================
*/

var effectInstances = make(map[string]PixelGenerator)
var globalConfig = new(GlobalEffectsConfig)
var validate *validator.Validate = validator.New()

func init() {
	validate.RegisterValidation("palette", validatePalette)
	validate.RegisterValidation("color", validateColor)
	// set global effect settings to default values
	if err := defaults.Set(&globalConfig); err != nil {
		panic(err)
	}
	// validate global effect settings
	if err := validate.Struct(&globalConfig); err != nil {
		log.Fatal(err)
	}
}

// Settings applied to all effects
type GlobalEffectsConfig struct {
	Brightness     float64 `json:"brightness" description:"Global brightness modifier" default:"1" validate:"gte=0,lte=1"`
	Hue            float64 `json:"hue" description:"Global hue modifier" default:"0" validate:"gte=0,lte=1"`
	Saturation     float64 `json:"saturation" description:"Global saturation modifier" default:"1" validate:"gte=0,lte=1"`
	TransitionMode string  `json:"transition_time" description:"Transition animation" default:"fade" validate:"oneof=fade wipe dissolve"` // TODO get this dynamically
	TransitionTime float64 `json:"transition_mode" description:"Duration of transitions (seconds)" default:"1" validate:"gte=0,lte=5"`
}

func validatePalette(fl validator.FieldLevel) bool {
	_, err := color.NewPalette(fl.Field().String())
	return err == nil
}

func validateColor(fl validator.FieldLevel) bool {
	_, err := color.NewColor(fl.Field().String())
	return err == nil
}

/*
Updates the global effect settings. Config can be given
as GlobalEffectsConfig, map[string]interface{}, or raw json
*/
func SetGlobalSettings(c interface{}) (err error) {
	var config GlobalEffectsConfig
	switch t := c.(type) {
	case GlobalEffectsConfig:
		config = c.(GlobalEffectsConfig)
	case map[string]interface{}:
		err = mapstructure.Decode(t, config)
	case []byte:
		err = json.Unmarshal(t, &config)
	default:
		err = fmt.Errorf("Invalid config type %T", c)
	}
	err = validate.Struct(&config)
	if err != nil {
		return err
	}
	globalConfig = &config
	return nil
}

// Get an existing pixel generator instance by its unique id
func Get(id string) (PixelGenerator, error) {
	if inst, exists := effectInstances[id]; exists {
		return inst, nil
	} else {
		return inst, fmt.Errorf("cannot retrieve effect of id: %s", id)
	}
}

// Kill an effect instance
func Destroy(id string) {
	if _, exists := effectInstances[id]; exists {
		delete(effectInstances, id)
	}
}

func JsonSchema() (jsonSchema []byte, err error) {
	schema, err := Schema()
	jsonSchema, err = utils.CreateJsonSchema(schema)
	return jsonSchema, err
}

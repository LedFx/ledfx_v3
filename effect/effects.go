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
	schema = make(map[string]interface{})
	// Copypaste for new effect types
	schema["energy"], err = utils.CreateSchema(reflect.TypeOf((*EnergyConfig)(nil)).Elem())
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
	err = effect.Initialize()
	if err != nil {
		return effect, nil
	}
	err = effect.UpdateConfig(config)
	return effect, err
}

/*
Nothing to modify below here =====================
*/

var effectInstances = make(map[string]PixelGenerator)
var globalConfig = GlobalEffectsConfig{}
var validate *validator.Validate = validator.New()

func init() {
	validate.RegisterValidation("palette", validatePalette)
	validate.RegisterValidation("color", validateColor)
	// set global effect settings to default values
	if err := defaults.Set(&globalConfig); err != nil {
		log.Fatal(err)
	}
	// validate global effect settings
	if err := validate.Struct(&globalConfig); err != nil {
		log.Fatal(err)
	}
}

// Settings applied to all effects
type GlobalEffectsConfig struct {
	Brightness     float64 `mapstructure:"brightness" json:"brightness" description:"Global brightness modifier" default:"1" validate:"gte=0,lte=1"`
	Hue            float64 `mapstructure:"hue" json:"hue" description:"Global hue modifier" default:"0" validate:"gte=0,lte=1"`
	Saturation     float64 `mapstructure:"saturation" json:"saturation" description:"Global saturation modifier" default:"1" validate:"gte=0,lte=1"`
	TransitionMode string  `mapstructure:"transition_time" json:"transition_time" description:"Transition animation" default:"fade" validate:"oneof=fade wipe dissolve"` // TODO get this dynamically
	TransitionTime float64 `mapstructure:"transition_mode" json:"transition_mode" description:"Duration of transitions (seconds)" default:"1" validate:"gte=0,lte=5"`
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
For incremental config updates, you must use map or json.
*/
func SetGlobalSettings(c interface{}) (err error) {
	// create a copy of the config
	newConfig := globalConfig
	// update values
	switch t := c.(type) {
	case GlobalEffectsConfig:
		newConfig = c.(GlobalEffectsConfig)
	case map[string]interface{}:
		err = mapstructure.Decode(t, &newConfig)
	case []byte:
		err = json.Unmarshal(t, &newConfig)
	default:
		err = fmt.Errorf("invalid config type %T", c)
	}
	if err != nil {
		return err
	}
	// validate new config
	err = validate.Struct(&newConfig)
	if err != nil {
		return err
	}
	// assign it ready for use
	globalConfig = newConfig
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
	delete(effectInstances, id)
}

func GetIDs() []string {
	ids := []string{}
	for id := range effectInstances {
		ids = append(ids, id)
	}
	return ids
}

func JsonSchema() (jsonSchema []byte, err error) {
	schema, err := Schema()
	if err != nil {
		return jsonSchema, err
	}
	jsonSchema, err = utils.CreateJsonSchema(schema)
	return jsonSchema, err
}

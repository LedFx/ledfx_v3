package effect

import (
	"encoding/json"
	"errors"
	"fmt"
	"ledfx/audio"
	"ledfx/color"
	"ledfx/config"
	"ledfx/logger"
	"ledfx/util"
	"log"
	"reflect"
	"strconv"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
)

/*
NEW EFFECTS MUST BE REGISTERED IN THIS SLICE AND THESE TWO FUNCTIONS =====================
*/

var effectTypes = []string{
	"energy",
	"palette",
	"fade",
	"weave",
	"pulse",
}

// Generate a map schema for all effects
func Schema() (schema map[string]interface{}, err error) {
	schema = make(map[string]interface{})
	schema["base"], err = util.CreateSchema(reflect.TypeOf((*BaseEffectConfig)(nil)).Elem())
	if err != nil {
		return schema, err
	}
	schema["types"] = effectTypes
	extraSchema := make(map[string]interface{})
	// Copypaste for new effect types, IF YOUR EFFECT HAS EXTRA CONFIG
	extraSchema["energy"], err = util.CreateSchema(reflect.TypeOf((*EnergyConfig)(nil)).Elem())
	if err != nil {
		return schema, err
	}
	schema["extra"] = extraSchema
	return schema, err
}

// Creates a new effect and returns its unique id.
// You can supply an ID. If an effect exists with this id, it will be destroyed and overwriten with this new effect
func New(new_id, effect_type string, pixelCount int, new_config interface{}) (effect *Effect, id string, err error) {
	switch effect_type {
	case "energy":
		effect = &Effect{
			pixelGenerator: &Energy{},
		}
	case "palette":
		effect = &Effect{
			pixelGenerator: &Palette{},
		}
	case "fade":
		effect = &Effect{
			pixelGenerator: &Fade{},
		}
	case "weave":
		effect = &Effect{
			pixelGenerator: &Weave{},
		}
	case "pulse":
		effect = &Effect{
			pixelGenerator: &Pulse{},
		}
	default:
		return effect, id, fmt.Errorf("'%s' is not a known effect type. Has it been registered in effects.go?", effect_type)
	}
	effect.Type = effect_type

	// if the id exists and has already been registered, overwrite the existing effect with that id
	if _, exists := effectInstances[new_id]; exists && new_id != "" {
		id = new_id
		Destroy(id)
		effectInstances[id] = effect
	} else { // otherwise, generate a new id
		for i := 0; ; i++ {
			id = effect_type + strconv.Itoa(i)
			_, exists := effectInstances[id]
			if !exists {
				effectInstances[id] = effect
				break
			}
		}
	}
	logger.Logger.WithField("context", "Effects").Debugf("Creating %s effect with id %s", effect_type, id)

	// initialise the new effect with its id and config
	effect.initialize(id, pixelCount)
	// Set effect's config to defaults
	if err = defaults.Set(&effect.Config); err != nil {
		Destroy(id)
		return effect, id, err
	}
	// update with any given config
	if err = effect.UpdateBaseConfig(new_config); err != nil {
		logger.Logger.WithField("context", "Effects").Warnf("Effect %s created with invalid config - aborting", id)
		Destroy(id)
		return effect, id, err
	}
	logger.Logger.WithField("context", "Effects").Infof("Created effect with id %s", id)
	return effect, id, err
}

/*
Nothing to modify below here =====================
*/

var effectInstances = make(map[string]*Effect)
var globalConfig = BaseEffectConfig{}
var validate *validator.Validate = validator.New()

func init() {
	err := validate.RegisterValidation("palette", validatePalette)
	if err != nil {
		log.Fatal(err)
	}
	err = validate.RegisterValidation("color", validateColor)
	if err != nil {
		log.Fatal(err)
	}
	// set global effect settings to default values
	if err = defaults.Set(&globalConfig); err != nil {
		log.Fatal(err)
	}
	// apply stored global config
	mapstructure.Decode(config.GetEffectsGlobal(), &globalConfig)
	// validate global effect settings
	if err = validate.Struct(&globalConfig); err != nil {
		log.Fatal(err)
	}
}

type TransitionConfig struct {
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
as BaseEffectConfig, map[string]interface{}, or raw json
For incremental config updates, you must use map or json.
*/
func SetGlobalSettings(c interface{}) (err error) {
	// create a copy of the config
	newConfig := globalConfig
	// update values
	switch t := c.(type) {
	case BaseEffectConfig:
		newConfig = c.(BaseEffectConfig)
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
	// validate all values
	if errs, ok := validate.Struct(&newConfig).(validator.ValidationErrors); ok {
		if errs != nil {
			errString := "Validation Errors: "
			for _, err := range errs {
				errString += fmt.Sprintf("Field %s with value %v; ", err.Field(), err.Value())
			}
			return errors.New(errString)
		}
	}
	// knowing that it's valid, pass it on to all the effects
	for _, e := range effectInstances {
		// we'll do this manually rather than calling updateBaseConfig to avoid unnecessary config saves and validation
		e.updateStoredProperties(newConfig)
		e.Config = newConfig
	}

	// assign it
	globalConfig = newConfig

	// TODO this could be changed so that global config follows incremental updates.
	// I think this way encourages users to use the global settings.
	// Globals will be applied as defaults to new effects
	configMap := make(map[string]interface{})
	err = mapstructure.Decode(globalConfig, &configMap)
	if err == nil {
		config.SetGlobalEffects(configMap)
	}
	return err
}

// Get an existing pixel generator instance by its unique id
func Get(id string) (*Effect, error) {
	if inst, exists := effectInstances[id]; exists {
		return inst, nil
	} else {
		return inst, fmt.Errorf("cannot retrieve effect of id: %s", id)
	}
}

// Kill an effect instance
func Destroy(id string) {
	logger.Logger.WithField("context", "Effects").Debugf("Deleting effect with id %s", id)
	audio.Analyzer.DeleteMelbank(id)
	config.DeleteEntry(config.Effect, id)
	delete(effectInstances, id)
	logger.Logger.WithField("context", "Effects").Infof("Deleted effect with id %s", id)
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
	jsonSchema, err = util.CreateJsonSchema(schema)
	return jsonSchema, err
}

func LoadFromConfig() error {
	storedEffects := config.GetEffects()
	for id, entry := range storedEffects {
		_, _, err := New(id, entry.Type, 100, entry.BaseConfig)
		if err != nil {
			return err
		}
	}
	return nil
}

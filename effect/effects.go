package effect

import (
	"errors"
	"fmt"
	"ledfx/audio"
	"ledfx/color"
	"ledfx/config"
	"ledfx/event"
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
NEW EFFECTS MUST BE REGISTERED IN THIS SLICE AND FUNCTION =====================
*/

var effectTypes = []string{
	"energy",
	"palette",
	"fade",
	"weave",
	"pulse",
	"strobe",
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
	case "strobe":
		effect = &Effect{
			pixelGenerator: &Strobe{},
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
func SetGlobalSettings(c map[string]interface{}) (err error) {
	// create a copy of the config
	newConfig := globalConfig
	err = mapstructure.Decode(c, &newConfig)
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
		// update effect configs incrementally with global config settings
		eConfig := e.Config
		err = mapstructure.Decode(&c, &eConfig)
		if err != nil {
			return err
		}
		e.updateStoredProperties(eConfig)
		e.Config = eConfig
		// manually invoke event
		event.Invoke(event.EffectUpdate,
			map[string]interface{}{
				"id":          e.ID,
				"base_config": eConfig,
			})
	}

	// save to config
	err = mapstructure.Decode(&newConfig, &c)
	if err == nil {
		config.SetGlobalEffects(c)
	}
	// invoke event
	event.Invoke(event.GlobalEffectUpdate,
		map[string]interface{}{
			"config": c,
		})
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
	audio.Analyzer.DeleteMelbank(id)
	config.DeleteEntry(config.Effect, id)
	delete(effectInstances, id)
	logger.Logger.WithField("context", "Effects").Infof("Deleted effect with id %s", id)
	// invoke event
	event.Invoke(event.EffectDelete,
		map[string]interface{}{
			"id": id,
		})
}

func GetIDs() []string {
	ids := []string{}
	for id := range effectInstances {
		ids = append(ids, id)
	}
	return ids
}

// Generate a map schema for all effects
func Schema() (schema map[string]interface{}, err error) {
	schema = make(map[string]interface{})
	schema["base"], err = util.CreateSchema(reflect.TypeOf((*BaseEffectConfig)(nil)).Elem())
	if err != nil {
		return schema, err
	}
	schema["types"] = effectTypes
	return schema, err
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

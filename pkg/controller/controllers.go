package controller

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/LedFx/ledfx/pkg/config"
	"github.com/LedFx/ledfx/pkg/event"
	"github.com/LedFx/ledfx/pkg/logger"
	"github.com/LedFx/ledfx/pkg/util"

	"github.com/go-playground/validator/v10"
)

// Creates a new controller and returns its unique id
func New(new_id string, config map[string]interface{}) (controller *Controller, id string, err error) {
	controller = &Controller{}

	// if the id exists and has already been registered, overwrite the existing controller with that id
	if _, exists := controllerInstances[new_id]; exists && new_id != "" {
		id = new_id
		Destroy(id)
		controllerInstances[id] = controller
	} else { // otherwise, generate a new id
		for i := 0; ; i++ {
			id = "controller" + strconv.Itoa(i)
			_, exists := controllerInstances[id]
			if !exists {
				controllerInstances[id] = controller
				break
			}
		}
	}
	logger.Logger.WithField("context", "Controllers").Debugf("Creating controller with id %s", id)

	// initialise the new controller with its id and config
	if err = controller.Initialize(id, config); err != nil {
		Destroy(id)
	}
	logger.Logger.WithField("context", "Controllers").Infof("Created controller with id %s", id)
	return controller, id, err
}

var controllerInstances = make(map[string]*Controller)

var validate *validator.Validate = validator.New()

// Get an existing controller instance by its unique id
func Get(id string) (*Controller, error) {
	if inst, exists := controllerInstances[id]; exists {
		return inst, nil
	} else {
		return inst, fmt.Errorf("cannot retrieve controller of id: %s", id)
	}
}

// Kill a controller instance
func Destroy(id string) {
	v, ok := controllerInstances[id]
	if !ok {
		logger.Logger.WithField("context", "Controllers").Warnf("Cannot delete %s, it doesn't exist", id)
		return
	}
	if v.State {
		v.Stop()
	}
	// remove it from saved states
	config.SetStates(GetStates())
	// disconnect any effects and devices
	if v.Effect != nil {
		DisconnectEffect(v.Effect.ID, v.ID)
	}
	for id := range v.Devices {
		DisconnectDevice(id, v.ID)
	}
	// remove it from config
	config.DeleteEntry(config.Controller, id)
	delete(controllerInstances, id)
	logger.Logger.WithField("context", "Controllers").Infof("Deleted %s", id)
	event.Invoke(event.ControllerDelete,
		map[string]interface{}{
			"id": id,
		})
}

// get activity status of all controllers
func GetStates() map[string]bool {
	states := make(map[string]bool)
	for _, v := range controllerInstances {
		states[v.ID] = v.State
	}
	return states
}

// set activity status of all controllers
func SetStates(states map[string]bool) (err error) {
	msg := ""
	for _, v := range controllerInstances {
		state, ok := states[v.ID]
		if !ok {
			continue
		}
		if v.State == state {
			continue
		}
		if state {
			err = v.Start()
		} else {
			v.Stop()
		}
		if err != nil {
			msg += err.Error()
		}
	}
	if msg != "" {
		return errors.New(msg)
	}
	config.SetStates(GetStates())
	return nil
}

func GetIDs() []string {
	ids := []string{}
	for id := range controllerInstances {
		ids = append(ids, id)
	}
	return ids
}

func LoadFromConfig() error {
	storedControllers := config.GetControllers()
	for id, entry := range storedControllers {
		_, _, err := New(id, entry.Config)
		if err != nil {
			return err
		}
	}
	return nil
}

func LoadStatesFromConfig() {
	SetStates(config.GetStates())
}

// Generate a map schema for all controllers
func Schema() (schema map[string]interface{}, err error) {
	schema, err = util.CreateSchema(reflect.TypeOf((*config.ControllerConfig)(nil)).Elem())
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

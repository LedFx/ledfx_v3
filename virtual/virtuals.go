package virtual

import (
	"errors"
	"fmt"
	"ledfx/config"
	"ledfx/logger"
	"ledfx/util"
	"reflect"
	"strconv"

	"github.com/go-playground/validator/v10"
)

// Creates a new virtual and returns its unique id
func New(new_id string, config map[string]interface{}) (virtual *Virtual, id string, err error) {
	virtual = &Virtual{}

	// if the id exists and has already been registered, overwrite the existing virtual with that id
	if _, exists := virtualInstances[new_id]; exists && new_id != "" {
		id = new_id
		Destroy(id)
		virtualInstances[id] = virtual
	} else { // otherwise, generate a new id
		for i := 0; ; i++ {
			id = "virtual" + strconv.Itoa(i)
			_, exists := virtualInstances[id]
			if !exists {
				virtualInstances[id] = virtual
				break
			}
		}
	}
	logger.Logger.WithField("context", "Virtuals").Debugf("Creating virtual with id %s", id)

	// initialise the new virtual with its id and config
	if err = virtual.Initialize(id, config); err != nil {
		Destroy(id)
	}
	return virtual, id, err
}

var virtualInstances = make(map[string]*Virtual)

var validate *validator.Validate = validator.New()

// Get an existing virtual instance by its unique id
func Get(id string) (*Virtual, error) {
	if inst, exists := virtualInstances[id]; exists {
		return inst, nil
	} else {
		return inst, fmt.Errorf("cannot retrieve virtual of id: %s", id)
	}
}

// Kill a virtual instance
func Destroy(id string) {
	delete(virtualInstances, id)
}

// get activity status of all virtuals
func GetStates() map[string]bool {
	states := make(map[string]bool)
	for _, v := range virtualInstances {
		states[v.ID] = v.Active
	}
	return states
}

// set activity status of all virtuals
func SetStates(states map[string]bool) (err error) {
	msg := ""
	for _, v := range virtualInstances {
		state, ok := states[v.ID]
		if !ok {
			continue
		}
		if v.Active == state {
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
	for id := range virtualInstances {
		ids = append(ids, id)
	}
	return ids
}

func LoadFromConfig() error {
	storedVirtuals := config.GetVirtuals()
	for id, entry := range storedVirtuals {
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

// Generate a map schema for all virtuals
func Schema() (schema map[string]interface{}, err error) {
	schema, err = util.CreateSchema(reflect.TypeOf((*config.VirtualConfig)(nil)).Elem())
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

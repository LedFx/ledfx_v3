package device

import (
	"fmt"
	"ledfx/config"
	"ledfx/logger"
	"ledfx/util"
	"reflect"
	"strconv"

	"github.com/go-playground/validator/v10"
)

var deviceTypes = []string{
	"udp",
}

// Creates a new device and returns its unique id
func New(new_id, device_type string, baseConfig map[string]interface{}, implConfig map[string]interface{}) (device *Device, id string, err error) {
	switch device_type {
	case "udp":
		device = &Device{
			pixelPusher: &UDP{},
		}
	default:
		return device, id, fmt.Errorf("%s is not a known device type", device_type)
	}
	device.Type = device_type

	// if the id exists and has already been registered, overwrite the existing device with that id
	if _, exists := deviceInstances[new_id]; exists && new_id != "" {
		id = new_id
		Destroy(id)
		deviceInstances[id] = device
	} else { // otherwise, generate a new id
		for i := 0; ; i++ {
			id = device_type + strconv.Itoa(i)
			_, exists := deviceInstances[id]
			if !exists {
				deviceInstances[id] = device
				break
			}
		}
	}
	logger.Logger.WithField("context", "Devices").Debugf("Creating %s device with id %s", device_type, id)

	// initialise the new device with its id and config
	if err = device.Initialize(id, baseConfig, implConfig); err != nil {
		Destroy(id)
	}
	return device, id, err
}

var deviceInstances = make(map[string]*Device)

var validate *validator.Validate = validator.New()

// Get an existing device instance by its unique id
func Get(id string) (*Device, error) {
	if inst, exists := deviceInstances[id]; exists {
		return inst, nil
	} else {
		return inst, fmt.Errorf("cannot retrieve device of id: %s", id)
	}
}

// Kill a device instance
func Destroy(id string) {
	delete(deviceInstances, id)
}

func GetIDs() []string {
	ids := []string{}
	for id := range deviceInstances {
		ids = append(ids, id)
	}
	return ids
}

func LoadFromConfig() error {
	storedDevices := config.GetDevices()
	for id, entry := range storedDevices {
		_, _, err := New(id, entry.Type, entry.BaseConfig, entry.ImplConfig)
		if err != nil {
			return err
		}
	}
	return nil
}

// Generate a map schema for all devices
func Schema() (schema map[string]interface{}, err error) {
	schema = make(map[string]interface{})
	schema["base"], err = util.CreateSchema(reflect.TypeOf((*config.BaseDeviceConfig)(nil)).Elem())
	if err != nil {
		return schema, err
	}
	schema["types"] = deviceTypes
	implSchema := make(map[string]interface{})
	// Copypaste for new effect types, IF YOUR EFFECT HAS EXTRA CONFIG
	implSchema["udp"], err = util.CreateSchema(reflect.TypeOf((*UDPConfig)(nil)).Elem())
	if err != nil {
		return schema, err
	}
	schema["impl"] = implSchema
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

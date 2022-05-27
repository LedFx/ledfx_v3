package device

import (
	"fmt"
	"strconv"
)

/*
NEW DEVICES MUST BE REGISTERED IN THESE TWO FUNCTIONS =====================
*/

// Generate a map schema for all devices
// func Schema() (schema map[string]interface{}, err error) {
// 	schema = make(map[string]interface{})
// 	schema["base"], err = utils.CreateSchema(reflect.TypeOf((*BaseEffectConfig)(nil)).Elem())
// 	if err != nil {
// 		return schema, err
// 	}
// 	extraSchema := make(map[string]interface{})
// 	// Copypaste for new device types
// 	extraSchema["energy"], err = utils.CreateSchema(reflect.TypeOf((*EnergyConfig)(nil)).Elem())
// 	if err != nil {
// 		return schema, err
// 	}
// 	schema["extra"] = extraSchema
// 	return schema, err
// }

// Creates a new device and returns its unique id
func New(device_type string, baseConfig BaseDeviceConfig, implConfig interface{}) (device *Device, id string, err error) {
	switch device_type {
	case "udp":
		device = &Device{
			pixelPusher: &UDP{},
		}
	default:
		return device, id, fmt.Errorf("%s is not a known device type", device_type)
	}

	// create an id and add it to the internal list of instances
	id = device_type
	for i := 0; ; i++ {
		id = device_type + strconv.Itoa(i)
		_, exists := deviceInstances[id]
		if !exists {
			deviceInstances[id] = device
			break
		}
	}
	// initialise the new device with its id and config
	if err = device.Initialize(id, baseConfig, implConfig); err != nil {
		return device, id, nil
	}
	// err = device.UpdateBaseConfig(config)
	return device, id, err
}

/*
Nothing to modify below here =====================
*/

var deviceInstances = make(map[string]*Device)

// var validate *validator.Validate = validator.New()

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

// func JsonSchema() (jsonSchema []byte, err error) {
// 	schema, err := Schema()
// 	if err != nil {
// 		return jsonSchema, err
// 	}
// 	jsonSchema, err = utils.CreateJsonSchema(schema)
// 	return jsonSchema, err
// }

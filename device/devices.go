package device

import (
	"fmt"
	"ledfx/config"
	"ledfx/event"
	"ledfx/logger"
	"ledfx/util"
	"reflect"
	"strconv"

	"github.com/go-playground/validator/v10"
)

type deviceInfo struct {
	Name      string   `mapstructure:"name" json:"name"`
	Info      string   `mapstructure:"info" json:"info"`
	Protocols []string `mapstructure:"protocols" json:"protocols"`
}

var deviceTypes = map[string]deviceInfo{
	"udp_stream": {
		Name:      "UDP Stream",
		Info:      "Stream pixel data over the network. Recommended for WLED",
		Protocols: []string{"WARLS", "DRGB", "DRGBW", "DNRGB", "DDP"},
	},
	"usb_serial": {
		Name:      "USB Serial",
		Info:      "Send pixel data to devices connected by USB. Works with WLED.",
		Protocols: []string{"Adalight", "TPM2"},
	},
	"artnet": {
		Name:      "ArtNet",
		Info:      "Implements ArtNet IV to send DMX-512 data over the network.",
		Protocols: []string{"ArtNet"},
	},
	"e131": {
		Name:      "E1.31 sACN",
		Info:      "Implements E1.31 sACN to send DMX-512 data over the network. Industry standard for lighting and control.",
		Protocols: []string{"E131"},
	},
}

// Creates a new device and returns its unique id
func New(new_id, device_type string, baseConfig map[string]interface{}, implConfig map[string]interface{}) (device *Device, id string, err error) {
	switch device_type {
	case "udp_stream":
		device = &Device{
			pixelPusher: &UDP{},
		}
	case "usb_serial":
		device = &Device{
			pixelPusher: &Serial{},
		}
	case "artnet":
		device = &Device{
			pixelPusher: &ArtNet{},
		}
	case "e131_sacn":
		device = &Device{
			pixelPusher: &E131{},
		}
	default:
		return device, id, fmt.Errorf("%s is not a known device type", device_type)
	}
	device.Type = device_type

	// if the id exists and has already been registered, overwrite the existing device with that id
	var prev_state State = Disconnected
	if old_d, exists := deviceInstances[new_id]; exists && new_id != "" {
		// save the state so we can restore it
		prev_state = old_d.State
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
	// restore its state
	if prev_state == Connected || prev_state == Connecting {
		go device.Connect()
	}
	logger.Logger.WithField("context", "Devices").Infof("Created device with id %s", id)
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
	if deviceInstances[id].State == Connected {
		deviceInstances[id].Disconnect()
	}
	config.DeleteEntry(config.Device, id)
	delete(deviceInstances, id)
	logger.Logger.WithField("context", "Devices").Infof("Deleted device with id %s", id)
	// invoke event
	event.Invoke(event.DeviceDelete,
		map[string]interface{}{
			"id": id,
		})
}

func GetIDs() []string {
	ids := []string{}
	for id := range deviceInstances {
		ids = append(ids, id)
	}
	return ids
}

func GetStates() map[string]State {
	states := map[string]State{}
	for _, d := range deviceInstances {
		states[d.ID] = d.State
	}
	return states
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
	implSchema["UDP Stream"], err = util.CreateSchema(reflect.TypeOf((*UDPConfig)(nil)).Elem())
	if err != nil {
		return schema, err
	}
	implSchema["USB Serial"], err = util.CreateSchema(reflect.TypeOf((*SerialConfig)(nil)).Elem())
	if err != nil {
		return schema, err
	}
	implSchema["ArtNet"], err = util.CreateSchema(reflect.TypeOf((*ArtNetConfig)(nil)).Elem())
	if err != nil {
		return schema, err
	}
	implSchema["E1.31 sACN"], err = util.CreateSchema(reflect.TypeOf((*E131Config)(nil)).Elem())
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

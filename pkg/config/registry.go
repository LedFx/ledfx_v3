package config

// import (
// 	"fmt"
// 	"reflect"
// 	"strconv"
// )

// /*
// A registry manages the creation and access of objects.
// */

// type RegistryEntry struct {
// 	Id     string
// 	Type   string
// 	Config ConfigSchema
// }

// type Registry struct {
// 	BaseType  reflect.Type             // eg. Effect
// 	Types     map[string]reflect.Type  // maps type strings to reflect types. {"energy": "EnergyEffect"}
// 	instances map[string]RegistryEntry // stores instances by mapping id strings to instances.
// }

// func (r *Registry) Create(type_str string, config ConfigSchema) {
// 	new_inst := reflect.New(r.Types[type_str]).Elem()
// 	// Create a unique id by appending a number to the type string
// 	i := 0
// 	for {
// 		id := type_str + strconv.Itoa(i)
// 		_, exists := r.instances[id]
// 		if !exists {
// 			r.instances[id] = new_inst
// 			break
// 		} else {
// 			i++
// 		}
// 	}
// 	// Apply the config to the new instance
// 	new_inst.Config
// }

// func (r *Registry) Get(id string) (inst RegistryEntry, err error) {
// 	if inst, exists := r.instances[id]; exists {
// 		return inst, err
// 	} else {
// 		return inst, fmt.Errorf("cannot retrieve object of id: %s", id)
// 	}
// }

// func (r *Registry) Destroy(id string) {
// 	if _, exists := r.instances[id]; exists {
// 		delete(r.instances, id)
// 	}
// }

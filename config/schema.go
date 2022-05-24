package config

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
)

/*
Example of config schema utilities which can:
 - Create a JSON schema which describes data types and allowed values for a struct
 - Validate incoming JSON to assign to a struct
 - Support embedded structs
*/

type ConfigSchemer interface {
	ValidateTags()
	Shema()
	JsonSchema()
}

type Student struct {
	Person
	Subject string  `json:"subject" description:"What they study" validate:"required"`
	GPA     float64 `json:"gpa" description:"Grade point average" validate:"required,gte=0,lte=4"`
}

type Person struct {
	Age        int      `json:"age" description:"Age of the person" validate:"required,gte=0,lte=130"`
	Name       string   `json:"name" description:"Name of the person" validate:"required"`
	FavNum     float64  `json:"favnum" description:"Favorite Number" validate:"gte=0" default:"3.14"`
	LikesPizza bool     `json:"likes_pizza" description:"Do they like pizza?" default:"true" validate:""`
	Friends    []string `json:"friends" description:"Names of friends" validate:"" default:"[]"`
}

func main() {
	// create json schema for student
	// this describes the values we expect for a "Student"
	studentType := reflect.TypeOf((*Student)(nil)).Elem()
	jsonSchema, err := CreateJsonSchema(studentType)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(jsonSchema))

	// The server used the schema and gave us this back
	objectString := `{
		"name": "Jake",
		"age": 20,
		"favnum": 16,
		"subject": "Computer Science",
		"gpa": 3.4
	}`
	student := NewStudent()

	// unmarshal and update values from json
	err = json.Unmarshal([]byte(objectString), &student)
	if err != nil {
		log.Fatal(err)
	}

	// validate all values
	validate := validator.New()
	err = validate.Struct(&student)
	if err != nil {
		log.Fatal(err)
	}
}

// create a new student, initialised with defaults
func NewStudent() Student {
	// create nil-initialised struct
	student := Student{}
	// set defaults
	if err := defaults.Set(&student); err != nil {
		panic(err)
	}

	return student
}

// create a new person, initialised with defaults
func NewPerson() Person {
	// create nil-initialised struct
	person := Person{}
	// set defaults
	if err := defaults.Set(&person); err != nil {
		panic(err)
	}

	return person
}

// Takes a type and checks that a schema can be made from its tags.
// Should be run at the start of the program to check all config schemas have valid tags.
func CheckConfigTags(t reflect.Type) error {
	fields := DeepFields(t)
	for _, f := range fields {
		_, hasDesc := f.Tag.Lookup("description") // this is the "description"
		_, hasDef := f.Tag.Lookup("default")      // this is the "default" field
		val, hasVal := f.Tag.Lookup("validate")   // this is the validator
		req := strings.Contains(val, "required")  // this is the "required" field
		if !hasVal {                              // TODO not log fatal
			return fmt.Errorf("field %s has no validator", f.Name)
		}
		if !hasDesc {
			return fmt.Errorf("field %s has no description", f.Name)
		}
		if req && hasDef {
			return fmt.Errorf("required field %s must not provide a default value", f.Name)
		}
		if !req && !hasDef {
			return fmt.Errorf("optional field %s must provide a default value", f.Name)
		}
	}
	return nil
}

// Creates a JSON schema for a given type
func CreateJsonSchema(t reflect.Type) ([]byte, error) {
	schema, err := CreateSchema(t)
	if err != nil {
		log.Fatal(err)
	}
	jsonSchema, err := json.MarshalIndent(schema, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	return jsonSchema, err
}

// Creates a schema for a given type
func CreateSchema(t reflect.Type) (map[string]interface{}, error) {
	err := CheckConfigTags(t)
	if err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	schema := make(map[string]interface{})
	fields := DeepFields(t)
	for _, f := range fields {
		desc, _ := f.Tag.Lookup("description")   // this is the "description"
		def, hasDef := f.Tag.Lookup("default")   // this is the "default" field
		val, _ := f.Tag.Lookup("validate")       // this is the "validator"
		jsonKey, _ := f.Tag.Lookup("json")       // this is the json "name"
		req := strings.Contains(val, "required") // this is the "required" field
		dataType := f.Type.String()

		// fmt.Print("-----------\n")
		// fmt.Printf("Field: %s\t Tag: %s\n", f.Name, f.Tag)
		// fmt.Printf("Title: %s\n", ToTitle(f.Name))
		// fmt.Printf("JSON key: %s\n", jsonKey)
		// fmt.Printf("Validation tag: %s\n", val)
		// fmt.Printf("Required: %t\tDefault: %s\n", req, def)
		// fmt.Printf("Description: %s\n", desc)
		// fmt.Printf("Type: %s\n", dataType)

		schemaEntry := make(map[string]interface{})
		schemaEntry["title"] = ToTitle(f.Name)
		schemaEntry["description"] = desc
		schemaEntry["required"] = req
		schemaEntry["type"] = dataType
		// convert the "default" value to its appropriate type
		switch dataType {
		case "string":
			schemaEntry["default"] = def
		case "bool":
			var x bool
			if hasDef {
				x, err = strconv.ParseBool(def)
				if err != nil {
					log.Fatal(err)
				}
			}
			schemaEntry["default"] = x
		case "int":
			var x int
			if hasDef {
				x, err = strconv.Atoi(def)
				if err != nil {
					log.Fatal(err)
				}
			}
			schemaEntry["default"] = x
		case "float64":
			var x float64
			if hasDef {
				x, err = strconv.ParseFloat(def, 64)
				if err != nil {
					log.Fatal(err)
				}
			}
			schemaEntry["default"] = x
		default:
			log.Fatalf("unimplemented config data type: %s", dataType)
		}

		validation := make(map[string]interface{})
		validators := strings.Split(val, ",")
		for _, entry := range validators {
			// split validator into two parts (eg. "gte=2" -> gte, 2)
			split := strings.Split(entry, "=")
			tag := split[0]
			var value string
			if len(split) > 1 {
				value = split[1]
			}
			// match tag to a few allowed tags
			// TODO expand these to accomodate for more validation tags as necessary
			switch tag {
			case "required": // special tag, doesn't matter here
				continue
			case "": // no validation, ignore
				continue
			case "gte":
				x, err := strconv.Atoi(value)
				if err != nil {
					log.Fatal(err)
				}
				validation["min"] = x
			case "lte":
				x, err := strconv.Atoi(value)
				if err != nil {
					log.Fatal(err)
				}
				validation["max"] = x
			default:
				log.Fatalf("unimplemented validation tag: %s", tag)
			}
		}
		schemaEntry["validation"] = validation
		schema[jsonKey] = schemaEntry

	}
	return schema, nil
}

// Flattens embedded structs to inspect the fields
func DeepFields(t reflect.Type) []reflect.StructField {
	fields := make([]reflect.StructField, 0)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		switch field.Type.Kind() {
		case reflect.Struct:
			fields = append(fields, DeepFields(field.Type)...)
		default:
			fields = append(fields, field)
		}
	}

	return fields
}

// Converts camelcase to a title. Eg: "IconName" -> "Icon Name"
func ToTitle(str string) string {
	matchFirstCap := regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap := regexp.MustCompile("([a-z0-9])([A-Z])")
	title := matchFirstCap.ReplaceAllString(str, "${1} ${2}")
	return matchAllCap.ReplaceAllString(title, "${1} ${2}")
}

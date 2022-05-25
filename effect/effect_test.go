package effect

import (
	"ledfx/color"
	"strings"
	"testing"

	"github.com/u2takey/go-utils/json"
)

func TestSchema(t *testing.T) {
	_, err := Schema()
	if err != nil {
		t.Error(err)
	}
}

func TestJsonSchema(t *testing.T) {
	schema, err := JsonSchema()
	// t.Log(string(schema))
	if err != nil {
		t.Errorf("%v, %v", schema, err)
	}
}

func TestEffectBaseFunctions(t *testing.T) {
	// Make a new effect
	c := map[string]interface{}{
		"brightness": 0.3,
	}
	_, err := New("energy", c)
	if err != nil {
		t.Error(err)
	}

	// Try to make an effect of unknown type
	_, err = New("doesnt_exist", c)
	if err == nil {
		t.Error("Invalid effect type should return an error")
	}

	// Try to make an effect with invalid config value
	c["brightness"] = 5000
	_, err = New("energy", c)
	if err == nil {
		t.Error("Invalid config should return an error")
	}

	// try to make an effect with invalid config type
	n := "this is clearly not a config"
	_, err = New("energy", n)
	if err == nil {
		t.Error("Invalid config should return an error")
	}

	// Get IDs to retrieve the effect
	id := GetIDs()[0]
	effect, err := Get(id)
	if err != nil {
		t.Error(err)
	}

	// Try to get the ID from the effect
	id = effect.GetID()
	if !strings.HasPrefix(id, "energy") {
		t.Errorf("Got wrong id: %s", id)
	}

	// Try get an invalid ID
	_, err = Get("doesnt_exist")
	if err == nil {
		t.Error("Invalid id should return an error")
	}

	// Run the effect on some pixels
	p := make(color.Pixels, 100)
	effect.Render(p)

	// Try to update with an invalid json
	c["nonsense"] = "data" // unknown keys are discarded
	c["brightness"] = 1.2  // invalid values throw error
	j, err := json.Marshal(c)
	if err != nil {
		t.Error(err)
	}
	err = effect.UpdateConfig(j)
	if err == nil {
		t.Error("Invalid config should return an error")
	}

	// Delete the effect
	Destroy(id)
}

func TestGlobalEffectSettings(t *testing.T) {
	// test with complete GlobalEffectsConfig
	g := GlobalEffectsConfig{
		Brightness:     0.5,
		Hue:            0.5,
		Saturation:     1,
		TransitionMode: "wipe",
		TransitionTime: 4,
	}
	err := SetGlobalSettings(g)
	if err != nil {
		t.Error(err)
	}

	// test with incremental map[string]interface
	m := map[string]interface{}{
		"brightness": 0.3,
	}
	err = SetGlobalSettings(m)
	if err != nil {
		t.Error(err)
	}

	// test with incremental json
	j, err := json.Marshal(m)
	if err != nil {
		t.Error(err)
	}
	err = SetGlobalSettings(j)
	if err != nil {
		t.Error(err)
	}

	// test with invalid config value
	m = map[string]interface{}{
		"brightness": 1.3,
	}
	err = SetGlobalSettings(m)
	if err == nil {
		t.Error("Invalid config should return an error")
	}

	// test with invalid config key
	m = map[string]interface{}{
		"brightness": 1.3,
	}
	err = SetGlobalSettings(m)
	if err == nil {
		t.Error("Invalid config should return an error")
	}

	// test with invalid config type
	s := "this isn't a config"
	err = SetGlobalSettings(s)
	if err == nil {
		t.Error("Invalid config should return an error")
	}
}

// cases := []struct {
// 	q string
// 	a Color
// 	e bool
// }{
// 	{"#ffFf00", Color{1, 1, 0}, false},
// 	{"RGB(0,255, 0)", Color{0, 1, 0}, false},
// 	{"#fF0", Color{0, 0, 0}, true},
// 	{"rgb(-1,0,256)", Color{}, true},
// 	{"#efghij", Color{}, true},
// 	{"nonsense color", Color{}, true},
// 	{"red", Color{1, 0, 0}, false},
// }
// for _, c := range cases {
// 	guess, err := NewColor(c.q)
// 	if (c.a != guess) || (err == nil == c.e) { // if the answer is wrong, or the error value is unexpected
// 		t.Errorf("Failed to parse %s: expected (%v, %v) but got (%v, %v)", c.q, c.a, c.e, guess, err)
// 	}
// }

package effect

import (
	"encoding/json"
	"ledfx/color"
	"strings"
	"testing"
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
	_, id, err := New("", "energy", 100, c)
	if err != nil {
		t.Error(err)
	}
	// Send it to the shadow realm
	Destroy(id)
	// Make sure it's dead
	_, err = Get(id)
	if err == nil {
		t.Error("Get effect with invalid ID should error")
	}

	// Try to make an effect of unknown type
	_, _, err = New("", "doesnt_exist", 100, c)
	if err == nil {
		t.Error("Invalid effect type should return an error")
	}

	// Try to make an effect with invalid config value
	c["brightness"] = 5000
	_, _, err = New("", "energy", 100, c)
	if err == nil {
		t.Error("Invalid config should return an error")
	}

	// try to make an effect with invalid config type
	n := "this is clearly not a config"
	_, _, err = New("", "energy", 100, n)
	if err == nil {
		t.Error("Invalid config should return an error")
	}

	// Make a new effect
	effect, _, err := New("", "energy", 100, nil)
	if err != nil {
		t.Error(err)
	}
	// Test if we can get the ID from the effect
	id = effect.GetID()
	if !strings.HasPrefix(id, "energy") {
		t.Errorf("Got wrong id: %s", id)
	}
	// Test if we can get the effect by its ID
	_, err = Get(id)
	if err != nil {
		t.Error(err)
	}
	// Test if we can get all effect IDs (this literally cannot go tits up)
	_ = GetIDs()

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
	err = effect.UpdateBaseConfig(j)
	if err == nil {
		t.Error("Invalid config should return an error")
	}

	// Delete the effect
	Destroy(id)
}

func TestGlobalEffectSettings(t *testing.T) {
	// test with incremental map[string]interface
	m := map[string]interface{}{
		"brightness": 0.3,
	}
	err := SetGlobalSettings(m)
	if err != nil {
		t.Error(err)
	}

	// test with invalid config value
	m = map[string]interface{}{
		"brightness": 1.3,
	}
	err = SetGlobalSettings(m)
	if err == nil {
		t.Error("Invalid config values should return an error")
	}

	// test with invalid config key
	m = map[string]interface{}{
		"floopydoop": 1.3,
	}
	err = SetGlobalSettings(m)
	if err != nil {
		t.Error(err)
	}
}

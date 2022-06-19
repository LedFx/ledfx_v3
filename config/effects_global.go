package config

func GetEffectsGlobal() map[string]interface{} {
	return store.EffectsGlobal
}

func SetGlobalEffects(g map[string]interface{}) {
	mu.Lock()
	defer mu.Unlock()
	store.EffectsGlobal = g
	saveConfig()
}

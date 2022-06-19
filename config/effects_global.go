package config

func GetEffectsGlobal() map[string]interface{} {
	return store.EffectsGlobal
}

func SetGlobalEffects(g map[string]interface{}) {
	store.EffectsGlobal = g
	saveConfig()
}

package config

func GetConnections() (effects, devices map[string]string) {
	return store.ConnEffect, store.ConnDevice
}

func SetConnections(effects, devices map[string]string) {
	mu.Lock()
	defer mu.Unlock()
	store.ConnEffect = effects
	store.ConnDevice = devices
	saveConfig()
}

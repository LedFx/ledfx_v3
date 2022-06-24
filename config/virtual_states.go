package config

func GetStates() map[string]bool {
	return store.VirtStates
}

func SetStates(states map[string]bool) {
	mu.Lock()
	defer mu.Unlock()
	store.VirtStates = states
	saveConfig()
}

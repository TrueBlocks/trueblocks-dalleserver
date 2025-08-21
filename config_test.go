package main

import "sync"

// resetConfigForTest clears the cached singleton so env changes take effect for LoadConfig.
func resetConfigForTest() {
	loadConfigOnce = sync.Once{}
	cachedConfig = Config{}
}

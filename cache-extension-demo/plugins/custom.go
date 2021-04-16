package plugins

// Struct for storing customs cache
type CustomConfiguration struct {
	Region string
	Names  []string
}

// Struct for caching the information
type Custom struct {
	CacheData CacheData
}

var customCache = make(map[string]Custom)

// Fetch custom cache
func GetCustomCache(name string) string {
	var custom = customCache[name]

	// If expired or not available return empty else return from cache
	if custom.CacheData.Data == "" || IsExpired(custom.CacheData.CacheExpiry) {
		return ""
	} else {
		return customCache[name].CacheData.Data
	}
}

// Store custom cache
func StoreCustomCache(name string, value string) string {
	customCache[name] = Custom{
		CacheData: CacheData{
			Data:        value,
			CacheExpiry: GetCacheExpiry(),
		},
	}
	return "OK"
}

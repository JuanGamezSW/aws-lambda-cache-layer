// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package extension

import (
	"cache-extension-demo/plugins"
	"io/ioutil"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"
)

// Constants definition
const (
	Parameters               = "parameters"
	Dynamodb                 = "dynamodb"
	Secret                   = "secret"
	Custom                   = "custom"
	FileName                 = "/var/task/config.yaml"
	InitializeCacheOnStartup = "CACHE_EXTENSION_INIT_STARTUP"
)

// Struct for storing CacheConfiguration
type CacheConfig struct {
	Parameters []plugins.ParameterConfiguration
	Dynamodb   []plugins.DynamodbConfiguration
	Secret     []plugins.SecretConfiguration
}

var cacheConfig = CacheConfig{}

// Initialize cache and start the background process to refresh cache
func InitCacheExtensions() {
	// Read the cache config file
	data := LoadConfigFile()
	// Unmarshal the configuration to struct
	err := yaml.Unmarshal([]byte(data), &cacheConfig)
	if err != nil {
		println(plugins.PrintPrefix, "Error: ", err.Error())
	}

	// Initialize Cache
	InitCache()
}

// Initialize individual cache
func InitCache() {

	// Read Lambda env variable
	var initCache = os.Getenv(InitializeCacheOnStartup)
	var initCacheInBool = false
	if initCache != "" {
		cacheInBool, err := strconv.ParseBool(initCache)
		if err != nil {
			println(plugins.PrintPrefix, "Error while converting CACHE_EXTENSION_INIT_STARTUP env variable")
			panic(plugins.PrintPrefix + "Error while converting CACHE_EXTENSION_INIT_STARTUP env variable " +
				initCache)
		} else {
			initCacheInBool = cacheInBool
		}
	}

	// Initialize map and load data from individual services if "CACHE_EXTENSION_INIT_STARTUP" = true
	plugins.InitParameters(cacheConfig.Parameters, initCacheInBool)
	plugins.InitDynamodb(cacheConfig.Dynamodb, initCacheInBool)
	plugins.InitSecret(cacheConfig.Secret, initCacheInBool)
}

// Route request to corresponding cache handlers
func RouteCache(cacheType string, name string) string {
	switch cacheType {
	case Parameters:
		return plugins.GetParameterCache(name)
	case Dynamodb:
		return plugins.GetDynamodbCache(name)
	case Secret:
		return plugins.GetSecretCache(name)
	case Custom:
		return plugins.GetCustomCache(name)
	default:
		return ""
	}
}

func PutCache(cacheType string, name string, value string) string {
	switch cacheType {
	case Custom:
		return plugins.StoreCustomCache(name, value)
	default:
		return ""
	}
}

// Load the config file
func LoadConfigFile() string {
	data, err := ioutil.ReadFile(FileName)
	if err != nil {
		println(plugins.PrintPrefix, "Error while reading config file")
		panic(err)
	}

	return string(data)
}

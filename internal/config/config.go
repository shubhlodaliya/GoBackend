package config

import (
	"encoding/json"
	"os"
	"reflect"
	"sync"
)

type DatabaseConfig struct {
	Uri string `json:"uri"`
}

type AppConfig struct {
	Port     string         `json:"port"`
	Database DatabaseConfig `json:"database"`
}

var (
	mutex   sync.RWMutex
	path    = "./config.json"
	Default = AppConfig{
		Port: ":443",
	}
	Current = Default // This will be updated at load time
)

func Load() error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			Current = Default
			return Save()
		}
		return err
	}

	mutex.Lock()
	defer mutex.Unlock()

	loaded := AppConfig{}
	if err := json.Unmarshal(data, &loaded); err != nil {
		return err
	}

	Current = merge(Default, loaded)
	return nil
}

func Save() error {
	mutex.RLock()
	defer mutex.RUnlock()

	data, err := json.MarshalIndent(Current, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func merge[T any](base, override T) T {
	baseVal := reflect.ValueOf(&base).Elem()
	overrideVal := reflect.ValueOf(override)

	mergeRecursive(baseVal, overrideVal)
	return base
}

func mergeRecursive(dst, src reflect.Value) {
	for i := 0; i < src.NumField(); i++ {
		srcField := src.Field(i)
		dstField := dst.Field(i)

		switch srcField.Kind() {
		case reflect.Struct:
			mergeRecursive(dstField, srcField)
		default:
			// If zero value, skip (don't override base)
			if !isZeroValue(srcField) {
				dstField.Set(srcField)
			}
		}
	}
}

func isZeroValue(v reflect.Value) bool {
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}

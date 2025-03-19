package config

import (
	"fmt"
	nurl "net/url"
	"os"
	"path/filepath"
	"reflect"
)

var (
	ErrNilFile = fmt.Errorf("no file")
	ErrNoPath  = fmt.Errorf("no file path")
)

func Merge(new, old map[string]interface{}) map[string]interface{} {
	return mergeMaps(new, old)
}

func mergeMaps(new, old map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for key, value := range new {
		out[key] = replaceFunc(value)
	}

	for oldKey, oldValue := range old {
		if newValue, exists := out[oldKey]; exists {
			newValueMap, okNewValueMap := newValue.(map[string]interface{})
			oldValueMap, okOldValueMap := oldValue.(map[string]interface{})
			newValueArray, okNewValueArray := newValue.([]interface{})
			oldValueArray, okOldValueArray := oldValue.([]interface{})

			if okNewValueMap && okOldValueMap {
				out[oldKey] = mergeMaps(newValueMap, oldValueMap)
			} else if okNewValueArray && okOldValueArray {
				if len(newValueArray) > 0 {
					if len(oldValueArray) > 0 {
						newValueArray2Map, okNewValueArray2 := newValueArray[0].(map[string]interface{})
						if okNewValueArray2 {
							newMap := make([]map[string]interface{}, len(oldValueArray))
							for i, newValueArrayValue := range oldValueArray {
								if i < len(newValueArray) {
									newValueArray2Map = newValueArray[i].(map[string]interface{})
								}

								oldValueArray2Map, okOldValueArray2 := newValueArrayValue.(map[string]interface{})
								if okOldValueArray2 {
									newMap[i] = mergeMaps(newValueArray2Map, oldValueArray2Map)
								}
							}
							out[oldKey] = newMap
						} else {
							out[oldKey] = replaceFunc(oldValueArray)
						}

					} else {
						continue
					}
				}
			} else if reflect.TypeOf(newValue) == reflect.TypeOf(oldValue) {
				out[oldKey] = replaceFunc(oldValue)
			}
		}
	}

	return out
}

func replaceFunc(value interface{}) interface{} {
	if v2, ok2 := value.(string); ok2 {
		return replace(v2)
	} else if v3, ok3 := value.([]string); ok3 {
		for i, c3 := range v3 {
			v3[i] = replace(c3)
		}
		return v3
	} else {
		return value
	}
}

func parseURL(url string) (string, error) {
	u, err := nurl.Parse(url)
	if err != nil {
		return "", err
	}
	p := u.Opaque
	if len(p) == 0 {
		p = u.Host + u.Path
	}

	if len(p) == 0 {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		p = wd
	} else if p[0:1] != "/" {
		// make path absolute if relative
		abs, err := filepath.Abs(p)
		if err != nil {
			return "", err
		}
		p = abs
	}
	return p, nil
}

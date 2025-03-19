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
	delete(old, "force")
	delete(old, "version")
	return mergeMaps(new, old)
}

func mergeMaps(new, old map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for key, value := range new {
		out[key] = replaceFunc(value)
	}

	if old != nil && len(old) > 0 {
		for oldKey, oldValue := range old {
			fmt.Println(oldKey, oldValue)
			if newValue, exists := out[oldKey]; exists {
				fmt.Println("2", oldKey, oldValue)
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
								fmt.Printf("3 %T %s %+v", oldValueArray, oldKey, oldValueArray)
								out[oldKey] = replaceFunc(oldValue)
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
	} else {
		return mergeMaps(out, out)
	}

	return out
}

func replaceFunc(value interface{}) interface{} {
	if v2, ok2 := value.(string); ok2 {
		return replace(v2)
	} else if v3, ok3 := value.([]interface{}); ok3 {
		for i, c3 := range v3 {
			if c2, ok := c3.(string); ok {
				v3[i] = replace(c2)
			}
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

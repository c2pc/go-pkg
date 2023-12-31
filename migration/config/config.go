package config

func Merge(a, b map[string]interface{}) map[string]interface{} {
	base := mergeMaps(a, b)
	base = clearMaps(base, a)

	return base
}

func MergeWithComment(a, b map[string]interface{}, comment string) map[string]interface{} {
	base := mergeMaps(a, b)
	base = clearMapsWithComment(base, a, comment)

	return base
}

func mergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v2, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = mergeMaps(bv, v2)
					continue
				}
			}
			out[k] = v2
			continue
		}

		out[k] = v
	}

	return out
}

func clearMaps(c, d map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range c {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := d[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = clearMaps(v, bv)
				}
			}
			continue
		}

		if _, ok := d[k]; ok {
			out[k] = v
		}
	}

	return out
}

func clearMapsWithComment(c, d map[string]interface{}, comment string) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range c {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := d[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = clearMapsWithComment(v, bv, comment)
				}
			} else {
				out[comment+k] = clearMapsWithComment(v, map[string]interface{}{}, comment)
			}
			continue
		}

		if _, ok := d[k]; !ok {
			if _, ok := out[comment+k]; !ok {
				out[comment+k] = v
			}
		} else {
			out[k] = v
		}
	}

	return out
}

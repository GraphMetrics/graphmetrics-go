package internal

func Bool2Int(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

func KeysAndValuesToMap(keysAndValues ...interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	length := len(keysAndValues)
	if length%2 != 0 {
		return result
	}
	for i := 0; i < len(keysAndValues); i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			return result
		}
		result[key] = keysAndValues[i+1]
	}
	return result
}

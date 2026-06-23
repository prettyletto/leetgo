package generator

func toPascalCase(slug string) string {
	result := make([]byte, 0, len(slug))
	capitalize := true
	for i := 0; i < len(slug); i++ {
		c := slug[i]
		if c == '-' || c == '_' {
			capitalize = true
			continue
		}
		if capitalize && c >= 'a' && c <= 'z' {
			c -= 32
		}
		result = append(result, c)
		capitalize = false
	}
	return string(result)
}

func toCamelCase(slug string) string {
	result := make([]byte, 0, len(slug))
	capitalize := false
	for i := 0; i < len(slug); i++ {
		c := slug[i]
		if c == '-' || c == '_' {
			capitalize = true
			continue
		}
		if capitalize && c >= 'a' && c <= 'z' {
			c -= 32
		}
		result = append(result, c)
		capitalize = false
	}
	return string(result)
}

func toSnakeCase(slug string) string {
	result := make([]byte, 0, len(slug))
	for i := 0; i < len(slug); i++ {
		c := slug[i]
		if c == '-' {
			result = append(result, '_')
		} else {
			result = append(result, c)
		}
	}
	return string(result)
}

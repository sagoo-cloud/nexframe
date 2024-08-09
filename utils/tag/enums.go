package tag

import "github.com/sagoo-cloud/nexframe/utils/json"

var (
	// Type name => enums json.
	enumsMap = make(map[string]json.RawMessage)
)

// SetGlobalEnums sets the global enums into package.
// Note that this operation is not concurrent safety.
func SetGlobalEnums(enumsJson string) error {
	return json.Unmarshal([]byte(enumsJson), &enumsMap)
}

// GetGlobalEnums retrieves and returns the global enums.
func GetGlobalEnums() (string, error) {
	enumsBytes, err := json.Marshal(enumsMap)
	if err != nil {
		return "", err
	}
	return string(enumsBytes), nil
}

// GetEnumsByType retrieves and returns the stored enums json by type name.
func GetEnumsByType(typeName string) string {
	return string(enumsMap[typeName])
}

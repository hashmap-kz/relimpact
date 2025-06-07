package diffs

import (
	"fmt"
	"strings"
)

func DiffStructFields(path, typeName string, oldFields, newFields []string) {
	oldMap := make(map[string]string) // FieldName → Type
	newMap := make(map[string]string)

	// Parse "FieldName TypeString"
	for _, f := range oldFields {
		parts := strings.SplitN(f, " ", 2)
		if len(parts) == 2 {
			oldMap[parts[0]] = parts[1]
		}
	}
	for _, f := range newFields {
		parts := strings.SplitN(f, " ", 2)
		if len(parts) == 2 {
			newMap[parts[0]] = parts[1]
		}
	}

	// Detect added fields
	for name, newType := range newMap {
		oldType, existed := oldMap[name]
		if !existed {
			fmt.Printf("- Added Field `%s` in `%s.%s`: `%s`\n", name, path, typeName, newType)
		} else if oldType != newType {
			fmt.Printf("- Field `%s` in `%s.%s` changed type: `%s` → `%s`\n", name, path, typeName, oldType, newType)
		}
	}

	// Detect removed fields
	for name, oldType := range oldMap {
		if _, exists := newMap[name]; !exists {
			fmt.Printf("- Removed Field `%s` from `%s.%s`: `%s`\n", name, path, typeName, oldType)
		}
	}
}

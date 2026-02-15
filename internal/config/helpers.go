package config

import (
	"fmt"
	"strconv"
	"strings"
)

func asMap(value any) map[string]any {
	if value == nil {
		return nil
	}

	typed, ok := value.(map[string]any)
	if ok {
		return typed
	}

	typedInterface, ok := value.(map[string]interface{})
	if ok {
		result := make(map[string]any, len(typedInterface))
		for k, v := range typedInterface {
			result[k] = v
		}
		return result
	}

	return nil
}

func asString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	case []byte:
		return string(typed)
	default:
		return ""
	}
}

func asInt64(value any) (int64, bool) {
	switch typed := value.(type) {
	case int:
		return int64(typed), true
	case int8:
		return int64(typed), true
	case int16:
		return int64(typed), true
	case int32:
		return int64(typed), true
	case int64:
		return typed, true
	case uint:
		return int64(typed), true
	case uint8:
		return int64(typed), true
	case uint16:
		return int64(typed), true
	case uint32:
		return int64(typed), true
	case uint64:
		if typed > ^uint64(0)>>1 {
			return 0, false
		}
		return int64(typed), true
	case float32:
		return int64(typed), true
	case float64:
		return int64(typed), true
	case jsonNumber:
		parsed, err := typed.Int64()
		if err != nil {
			return 0, false
		}
		return parsed, true
	case string:
		parsed, err := strconv.ParseInt(strings.TrimSpace(typed), 10, 64)
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}

func shortAccountID(accountID string) string {
	trimmed := strings.TrimSpace(accountID)
	if trimmed == "" {
		return ""
	}
	if len(trimmed) <= 12 {
		return trimmed
	}
	return trimmed[:6] + "..." + trimmed[len(trimmed)-4:]
}

type jsonNumber interface {
	Int64() (int64, error)
}

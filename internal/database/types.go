package database

import (
	"encoding/json"
	"errors"
	"strings"
)

// FlexibleBool permite deserializar JSON booleano tanto si viene como bool (true/false) como int (1/0) o string ("true"/"false"/"1"/"0")
type FlexibleBool bool

func (b *FlexibleBool) UnmarshalJSON(data []byte) error {
	// 1. Bool directo
	var asBool bool
	if err := json.Unmarshal(data, &asBool); err == nil {
		*b = FlexibleBool(asBool)
		return nil
	}
	// 2. Número (0 o 1)
	var asInt int
	if err := json.Unmarshal(data, &asInt); err == nil {
		*b = FlexibleBool(asInt != 0)
		return nil
	}
	// 3. String
	var asStr string
	if err := json.Unmarshal(data, &asStr); err == nil {
		clean := strings.TrimSpace(strings.ToLower(asStr))
		*b = FlexibleBool(clean == "true" || clean == "1")
		return nil
	}
	return errors.New("valor booleano inválido")
}

func (b FlexibleBool) Bool() bool {
	return bool(b)
}

func (b *FlexibleBool) ValueOrDefault(def bool) bool {
	if b == nil {
		return def
	}
	return bool(*b)
}

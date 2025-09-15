package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func EncodeJson[T any](w http.ResponseWriter, r *http.Request, statusCode int, data T) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		return fmt.Errorf("erro ao fazer encode do json: %v", err)
	}

	return nil
}

func DecodeJson[T any](r *http.Request) (T, error) {
	var data T

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return data, fmt.Errorf("erro ao fazer decode do json: %v", err)
	}
	return data, nil
}

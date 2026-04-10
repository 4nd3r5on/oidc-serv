package postgres

import "encoding/json"

// marshalJSON serializes v to JSON, returning nil for a nil value or a JSON null result.
func marshalJSON(v any) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if string(b) == "null" {
		return nil, nil
	}
	return b, nil
}

// unmarshalJSON deserialises b into dest, treating empty input as a no-op.
func unmarshalJSON(b []byte, dest any) error {
	if len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, dest)
}

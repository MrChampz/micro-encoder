package utils

import "encoding/json"

func IsJson(str string) error {
	var json_ struct{}

	if err := json.Unmarshal([]byte(str), &json_); err != nil {
		return err
	}

	return nil
}

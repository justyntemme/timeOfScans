package util
import (
  "strings"
  "encoding/json"
  "errors"
  "github.com/araddon/dateparse"
)
func ExtractTimeValues(jsonResponse string) ([]string, error) {
	if jsonResponse == "" {
		return nil, errors.New("empty JSON response")
	}

	if strings.HasPrefix(strings.TrimSpace(jsonResponse), "<!doctype html>") {
		return nil, errors.New("received HTML response instead of JSON")
	}

	if strings.TrimSpace(jsonResponse) == "null" {
		return nil, errors.New("received null response")
	}

	var data []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonResponse), &data); err != nil {
		return nil, err
	}

	timeValues := make([]string, 0)
	for _, item := range data {
			if scanTimeStr, ok := item["scanTime"].(string); ok {
				timeValue, err := dateparse.ParseAny(scanTimeStr)
				if err != nil {
					return nil, err
				}
				formattedTime := timeValue.UTC().Format("03:04 PM")
				timeValues = append(timeValues, formattedTime)
			}
	}

	return timeValues, nil
}

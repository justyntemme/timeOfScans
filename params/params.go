package params

import (
	"fmt"
	"os"
)

func GetEnvVars(paramNames []string) (map[string]string, error) {
	missingParams := []string{}
	values := make(map[string]string)

	for _, paramName := range paramNames {
		paramValue := os.Getenv(paramName)
		if paramValue == "" {
			missingParams = append(missingParams, paramName)
		} else {
			values[paramName] = paramValue
		}
	}

	if len(missingParams) > 0 {
		return nil, fmt.Errorf("missing parameters: %v", missingParams)
	}

	return values, nil
}

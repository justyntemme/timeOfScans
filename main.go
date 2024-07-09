package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/araddon/dateparse"
  "github.com/justyntemme/timeOfScans/params"
)

// Global Variables
var tlUrl = os.Getenv("TL_URL")



func generateCwpToken(accessKey, accessSecret string) (string, error) {
	authUrl := fmt.Sprintf("%s/api/v1/authenticate", tlUrl)

	body := map[string]string{
		"username": accessKey,
		"password": accessSecret,
	}

	bodyJson, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", authUrl, bytes.NewBuffer(bodyJson))
	if err != nil {
		return "", err
	}

	req.Header.Set("accept", "application/json; charset=UTF-8")
	req.Header.Set("content-type", "application/json")

	client := &http.Client{Timeout: time.Second * 60}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unable to acquire token with error code: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	token, ok := result["token"].(string)
	if !ok {
		return "", errors.New("token not found in response")
	}

	return token, nil
}

func getScans(token string, offset, limit int) (int, string, error) {
	scanUrl := fmt.Sprintf("%s/api/v1/scans", tlUrl)
	req, err := http.NewRequest("GET", scanUrl, nil)
	if err != nil {
		return 0, "", err
	}

	req.Header.Set("accept", "application/json; charset=UTF-8")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	query := req.URL.Query()
	query.Add("offset", strconv.Itoa(offset))
	query.Add("limit", strconv.Itoa(limit))
	req.URL.RawQuery = query.Encode()

	client := &http.Client{Timeout: time.Second * 60}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, "", err
	}

	return resp.StatusCode, string(respBody), nil
}

func extractTimeValues(jsonResponse string) ([]string, error) {
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
		if entityInfo, ok := item["entityInfo"].(map[string]interface{}); ok {
			if scanTimeStr, ok := entityInfo["scanTime"].(string); ok {
				timeValue, err := dateparse.ParseAny(scanTimeStr)
				if err != nil {
					return nil, err
				}
				formattedTime := timeValue.UTC().Format("03:04 PM")
				timeValues = append(timeValues, formattedTime)
			}
		}
	}

	return timeValues, nil
}

func getAllScansWithTimeCounts(token string, limit int, result chan<- map[string]int, wg *sync.WaitGroup) {
	defer wg.Done()
	offset := 0
	timeCounts := make(map[string]int)

	for {
		statusCode, responseText, err := getScans(token, offset, limit)
		if err != nil || statusCode != http.StatusOK {
			fmt.Printf("Error fetching scans: %v, Status Code: %d\n", err, statusCode)
			break
		}

		timeValues, err := extractTimeValues(responseText)
		if err != nil {
			fmt.Printf("Error extracting time values: %v\n", err)
			break
		}

		for _, timeValue := range timeValues {
			timeCounts[timeValue]++
		}

		var data []map[string]interface{}
		if err := json.Unmarshal([]byte(responseText), &data); err != nil {
			fmt.Printf("Error decoding JSON response: %v\n", err)
			break
		}

		if len(data) < limit {
			break
		}

		offset += limit
	}

	result <- timeCounts
}

func main() {
	paramKeys := []string{"PC_IDENTITY", "PC_SECRET", "TL_URL"}
  paramValues, err := params.GetEnvVars(paramKeys)	
  if err != nil {
		fmt.Println(err)
		return
	}

		cwpToken, err := generateCwpToken(paramValues["PC_IDENTITY"], paramValues["PC_SECRET"])

	if err != nil {
		fmt.Println(err)
		return
	}

	var wg sync.WaitGroup
	result := make(chan map[string]int)

	wg.Add(1)
	go getAllScansWithTimeCounts(cwpToken, 5, result, &wg)

	go func() {
		wg.Wait()
		close(result)
	}()

	timeCounts := make(map[string]int)
	for counts := range result {
		for key, count := range counts {
			timeCounts[key] += count
		}
	}

	fmt.Println(timeCounts)
}

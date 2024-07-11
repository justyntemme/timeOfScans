package net

import (
	"encoding/json"
	"errors"
	"bytes"
	"strconv"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/justyntemme/timeOfScans/util"
)

func GetAllScansWithTimeCounts(token string, limit int, result chan<- map[string]int, wg *sync.WaitGroup, tlUrl string) {
	defer wg.Done()
	offset := 0
	timeCounts := make(map[string]int)
	rateLimiter := make(chan time.Time, 30)

	// Fill channel with initial rate limit tokens
	for i := 0; i < 30; i++ {
		rateLimiter <- time.Now()
	}

	// Token refiller
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for t := range ticker.C {
			for i := 0; i < 30; i++ {
				rateLimiter <- t
			}
		}
	}()

	for {
		<-rateLimiter

		statusCode, responseText, err := getScans(token, offset, limit, tlUrl)
		if err != nil || statusCode != http.StatusOK {
			fmt.Printf("Error fetching scans: %v, Status Code: %d\n", err, statusCode)
			break
		}

		timeValues, err := util.ExtractTimeValues(responseText)
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

func getScans(token string, offset, limit int, tlUrl string) (int, string, error) {
	scanUrl := fmt.Sprintf("%s/api/v1/images", tlUrl)
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

func GenerateCwpToken(accessKey, accessSecret string, tlUrl string) (string, error) {
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

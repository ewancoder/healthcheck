package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	uri := os.Getenv("HEALTHCHECK_URI")
	file := os.Getenv("HEALTHCHECK_FILE")
	if (uri == "" && file == "") || (uri != "" && file != "") {
		fmt.Println("Either HEALTHCHECK_URI or HEALTHCHECK_FILE must be set.")
		os.Exit(1)
	}

	err := error(nil)
	if uri != "" {
		fmt.Printf("Running URI healthcheck: %s\n", uri)

		timeoutStr := os.Getenv("HEALTHCHECK_URI_TIMEOUT")
		expectedStatusCodeStr := os.Getenv("HEALTHCHECK_URI_STATUS_CODE")

		timeout := 10
		if t, err := strconv.Atoi(timeoutStr); err == nil {
			timeout = t
			fmt.Printf("Using custom timeout: %ds\n", timeout)
		} else {
			fmt.Printf("Using default timeout: %ds\n", timeout)
		}

		expectedStatusCode := http.StatusOK
		if code, err := strconv.Atoi(expectedStatusCodeStr); err == nil {
			expectedStatusCode = code
			fmt.Printf("Using custom status code: %d\n", expectedStatusCode)
		} else {
			fmt.Printf("Using default status code: %d\n", expectedStatusCode)
		}

		err = uriHealthCheck(uri, timeout, expectedStatusCode)
		if err != nil {
			fmt.Printf("URI healthcheck failed: %v\n", err)
			os.Exit(1)
		}
	}

	if file != "" {
		fmt.Printf("Running file healthcheck: %s\n", file)

		maxAgeStr := os.Getenv("HEALTHCHECK_FILE_MAX_AGE")
		maxAge := 60
		if m, err := strconv.Atoi(maxAgeStr); err == nil {
			maxAge = m
			fmt.Printf("Using custom file max age: %ds\n", maxAge)
		} else {
			fmt.Printf("Using default file max age: %ds\n", maxAge)
		}

		err = fileHealthCheck(file, maxAge)
		if err != nil {
			fmt.Printf("File healthcheck failed: %v\n", err)
			os.Exit(1)
		}
	}
}

func uriHealthCheck(uri string, timeout int, expectedStatusCode int) error {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	resp, err := client.Get(uri)
	if err != nil {
		return fmt.Errorf("failed to get URI: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatusCode {
		return fmt.Errorf("invalid status code: %d, expected %d", resp.StatusCode, expectedStatusCode)
	}

	return nil
}

func fileHealthCheck(filepath string, maxAgeSeconds int) error {
	info, err := os.Stat(filepath)
	if err != nil {
		return fmt.Errorf("failed to access the file: %v", err)
	}

	if time.Since(info.ModTime()) > time.Duration(maxAgeSeconds)*time.Second {
		return fmt.Errorf("the file was not modified in the last %d seconds", maxAgeSeconds)
	}

	return nil
}

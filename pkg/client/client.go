package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aep-dev/aep-lib-go/pkg/api"
)

func Create(ctx context.Context, r *api.Resource, c *http.Client, serverUrl string, body map[string]interface{}, parameters map[string]string) (map[string]interface{}, error) {
	suffix := ""
	if r.CreateMethod != nil && r.CreateMethod.SupportsUserSettableCreate {
		id, ok := body["id"]
		if !ok {
			return nil, fmt.Errorf("id field not found in %v", body)
		}
		idString, ok := id.(string)
		if !ok {
			return nil, fmt.Errorf("id field is not string %v", id)
		}

		suffix = fmt.Sprintf("?id=%s", idString)
	}
	url := createBase(ctx, r, serverUrl, parameters, suffix)

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("error marshalling JSON: %v", err)
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, fmt.Errorf("error creating post request: %v", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}
	var data map[string]interface{}
	err = json.Unmarshal(respBody, &data)
	if err != nil {
		return nil, err
	}

	err = CheckErrors(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func Read(ctx context.Context, c *http.Client, serverUrl string, path string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/%s", serverUrl, strings.TrimPrefix(path, "/"))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating post request: %v", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	err = CheckErrors(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func Delete(ctx context.Context, c *http.Client, serverUrl string, path string) error {
	url := fmt.Sprintf("%s/%s", serverUrl, strings.TrimPrefix(path, "/"))

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("error creating delete request: %v", err)
	}

	resp, err := c.Do(req)
	var data map[string]interface{}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	err = CheckErrors(data)
	if err != nil {
		return err
	}
	return err
}

func Update(ctx context.Context, c *http.Client, serverUrl string, path string, body map[string]interface{}) error {
	url := fmt.Sprintf("%s/%s", serverUrl, strings.TrimPrefix(path, "/"))

	reqBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("error marshalling JSON for request body: %v", err)
	}

	req, err := http.NewRequest("PATCH", url, strings.NewReader(string(reqBody)))
	if err != nil {
		return fmt.Errorf("error creating delete request: %v", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return err
	}

	var data map[string]interface{}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}

	err = json.Unmarshal(respBody, &data)
	if err != nil {
		return fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	err = CheckErrors(data)
	if err != nil {
		return err
	}

	return CheckErrors(data)
}

func CheckErrors(resp map[string]interface{}) error {
	e, ok := resp["error"]
	if ok {
		return fmt.Errorf("returned errors, %v", e)
	}
}

func createBase(ctx context.Context, r *api.Resource, serverUrl string, parameters map[string]string, suffix string) string {
	serverUrl = strings.TrimSuffix(serverUrl, "/")
	urlElems := []string{serverUrl}
	for i, elem := range r.PatternElems {
		if i == len(r.PatternElems)-1 {
			continue
		}

		if i%2 == 0 {
			urlElems = append(urlElems, elem)
		} else {
			paramName := elem[1 : len(elem)-1]
			if value, ok := parameters[paramName]; ok {
				if strings.Contains(value, "/") {
					value = strings.Split(value, "/")[len(strings.Split(value, "/"))-1]
				}
				urlElems = append(urlElems, value)
			}
		}
	}
	result := strings.Join(urlElems, "/")
	if suffix != "" {
		result = result + suffix
	}
	return result
}

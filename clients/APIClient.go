package clients

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	microappCtx "github.com/islax/microapp/context"
	microappError "github.com/islax/microapp/error"
)

// APIClient represents the actual client calling microservice
type APIClient struct {
	AppName    string
	BaseURL    string
	HTTPClient *http.Client
}

// DoRequestWithResponseParam do request with response param
func (apiClient *APIClient) DoRequestWithResponseParam(context microappCtx.ExecutionContext, url string, requestMethod string, rawToken string, payload map[string]interface{}, out interface{}) error {
	apiURL := apiClient.BaseURL + url
	var body io.Reader
	if payload != nil {
		bytePayload, err := json.Marshal(payload)
		if err != nil {
			return microappError.NewAPICallError(apiURL, nil, nil, fmt.Errorf("Unable to marshal payload: %w", err))
		}
		body = bytes.NewBuffer(bytePayload)
	}

	request, err := http.NewRequest(requestMethod, apiURL, body)
	if err != nil {
		return microappError.NewAPICallError(apiURL, nil, nil, fmt.Errorf("Unable to create HTTP request: %w", err))
	}

	if rawToken != "" {
		if strings.HasPrefix(rawToken, "Bearer") {
			request.Header.Add("Authorization", rawToken)
		} else {
			request.Header.Add("Authorization", "Bearer "+rawToken)
		}
	}
	request.Header.Set("X-Client", apiClient.AppName)
	request.Header.Set("X-Correlation-ID", context.GetCorrelationID())
	request.Header.Set("Content-Type", "application/json")

	response, err := apiClient.HTTPClient.Do(request)
	if err != nil {
		return microappError.NewAPICallError(apiURL, nil, nil, fmt.Errorf("Unable to invoke API: %w", err))
	}

	defer response.Body.Close()
	if response.StatusCode > 300 { // All 3xx, 4xx, 5xx are considered errors
		responseBodyString := ""
		if responseBodyBytes, err := ioutil.ReadAll(response.Body); err != nil {
			responseBodyString = string(responseBodyBytes)
		}
		return microappError.NewAPICallError(apiURL, &response.StatusCode, &responseBodyString, fmt.Errorf("Received non-success code: %v", response.StatusCode))
	}

	if out != nil {
		err = json.NewDecoder(response.Body).Decode(out)
		if err != nil {
			return microappError.NewAPICallError(apiURL, &response.StatusCode, nil, fmt.Errorf("Unable parse response payload: %w", err))
		}
	}
	return nil
}

func (apiClient *APIClient) doRequest(context microappCtx.ExecutionContext, url string, requestMethod string, rawToken string, payload map[string]interface{}) (interface{}, error) {
	apiURL := apiClient.BaseURL + url
	var body io.Reader
	if payload != nil {
		bytePayload, err := json.Marshal(payload)
		if err != nil {
			return nil, microappError.NewAPICallError(apiURL, nil, nil, fmt.Errorf("Unable to marshal payload: %w", err))
		}
		body = bytes.NewBuffer(bytePayload)
	}

	request, err := http.NewRequest(requestMethod, apiURL, body)
	if err != nil {
		return nil, microappError.NewAPICallError(apiURL, nil, nil, fmt.Errorf("Unable to create HTTP request: %w", err))
	}

	if rawToken != "" {
		if strings.HasPrefix(rawToken, "Bearer") {
			request.Header.Add("Authorization", rawToken)
		} else {
			request.Header.Add("Authorization", "Bearer "+rawToken)
		}
	}
	request.Header.Set("X-Client", apiClient.AppName)
	request.Header.Set("X-Correlation-ID", context.GetCorrelationID())
	request.Header.Set("Content-Type", "application/json")

	response, err := apiClient.HTTPClient.Do(request)
	if err != nil {
		return nil, microappError.NewAPICallError(apiURL, nil, nil, fmt.Errorf("Unable to invoke API: %w", err))
	}

	defer response.Body.Close()
	if response.StatusCode > 300 { // All 3xx, 4xx, 5xx are considered errors
		responseBodyString := ""
		if responseBodyBytes, err := ioutil.ReadAll(response.Body); err != nil {
			responseBodyString = string(responseBodyBytes)
		}
		return nil, microappError.NewAPICallError(apiURL, &response.StatusCode, &responseBodyString, fmt.Errorf("Received non-success code: %v", response.StatusCode))
	}

	var mapResponse interface{}
	err = json.NewDecoder(response.Body).Decode(&mapResponse)
	if err != nil {
		return nil, microappError.NewAPICallError(apiURL, &response.StatusCode, nil, fmt.Errorf("Unable parse response payload: %w", err))
	}

	return mapResponse, nil
}

// DoGet is a generic method to carry out RESTful calls to the other external microservices in ISLA
func (apiClient *APIClient) DoGet(context microappCtx.ExecutionContext, requestString string, rawToken string) (map[string]interface{}, error) {
	response, err := apiClient.doRequest(context, requestString, http.MethodGet, rawToken, nil)
	if err != nil {
		return nil, err
	}

	mapResponse, ok := response.(map[string]interface{})
	if !ok {
		return nil, errors.New("Could not parse Json to map")
	}
	return mapResponse, nil
}

// DoGetList is a generic method to carry out RESTful calls to the other external microservices in ISLA
func (apiClient *APIClient) DoGetList(context microappCtx.ExecutionContext, requestString string, rawToken string) ([]map[string]interface{}, error) {
	response, err := apiClient.doRequest(context, requestString, http.MethodGet, rawToken, nil)
	if err != nil {
		return nil, err
	}
	sliceOfGenericObjects, ok := response.([]interface{})
	if !ok {
		return nil, errors.New("Could not parse Json to map")
	}
	var sliceOfMapObjects []map[string]interface{}
	for _, obj := range sliceOfGenericObjects {
		mapObject, ok := obj.(map[string]interface{})
		if ok {
			sliceOfMapObjects = append(sliceOfMapObjects, mapObject)
		} else {
			return nil, errors.New("Could not parse Json to map")
		}
	}
	return sliceOfMapObjects, nil
}

// DoPost is a generic method to carry out RESTful calls to the other external microservices in ISLA
func (apiClient *APIClient) DoPost(context microappCtx.ExecutionContext, requestString string, rawToken string, payload map[string]interface{}) (map[string]interface{}, error) {
	response, err := apiClient.doRequest(context, requestString, http.MethodPost, rawToken, payload)
	if err != nil {
		return nil, err
	}

	mapResponse, ok := response.(map[string]interface{})
	if !ok {
		return nil, errors.New("Could not parse Json to map")
	}
	return mapResponse, nil
}

// DoDelete is a generic method to carry out RESTful calls to the other external microservices in ISLA
func (apiClient *APIClient) DoDelete(context microappCtx.ExecutionContext, requestString string, rawToken string, payload map[string]interface{}) error {
	_, err := apiClient.doRequest(context, requestString, http.MethodDelete, rawToken, payload)
	if err != nil {
		return err
	}
	return nil
}

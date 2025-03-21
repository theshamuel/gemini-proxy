package service

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
)

// GeminiProxy represents proxy service
type GeminiProxy struct {
	OriginURL string
	Client    http.Client
	APIKey    string
	Lock      sync.Mutex
}

// Send request to Gemini API and proxy back the Gemini response
func (r *GeminiProxy) Send(request io.ReadCloser) ([]byte, error) {
	if r.APIKey == "" {
		return nil, fmt.Errorf("gemini API key is not found")
	}

	url := r.OriginURL + r.APIKey

	httpReq, err := http.NewRequest("POST", url, request)

	if err != nil {
		log.Printf("[ERROR] cannot create POST request: %#v;", err)
		return nil, err
	}

	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("Priority", "u=1, i")
	httpResp, err := r.Client.Do(httpReq)
	if err != nil {
		log.Printf("[ERROR] can not make POST request: %#v", err)
		if errClose := httpResp.Body.Close(); errClose != nil {
			log.Printf("[ERROR] can not close response body %#v", errClose)
		}
		return nil, err
	}

	if httpResp == nil {
		return nil, fmt.Errorf("response from Gemini is nil")
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response from Gemini is not 200: %s", httpResp.Status)
	}

	byteResp, err := io.ReadAll(httpResp.Body)
	if err != nil {
		log.Printf("[ERROR] can not read response body %#v", err)
		return nil, err
	}

	return byteResp, nil
}

func (r *GeminiProxy) GetMutex() *sync.Mutex {
	return &r.Lock
}

package service

// GeminiProxy represents proxy service
type GeminiProxy struct {
}

// GeminiProxyRequest represents request type for Min
type GeminiProxyRequest struct {
	data string `json:"data"`
}

// GeminiProxyResponse represents request type for Min
type GeminiProxyResponse struct {
	data string `json:"data"`
}

// Send request to Gemini API and proxy back the Gemini response
func (s *GeminiProxy) Send(request GeminiProxyRequest) (*GeminiProxyResponse, error) {
	return nil, nil
}

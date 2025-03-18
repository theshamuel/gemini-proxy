package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// GeminiProxy represents proxy service
type GeminiProxy struct {
	OriginURL string
	Client    http.Client
}

type GeminiOriginRequest struct {
	Contents []Content `json:"contents"`
}

// GeminiOriginResponse Gemini origin response
// Full structure is:
//
//	type GeminiResponse struct {
//		Candidate []struct {
//			Content struct {
//				Part []struct {
//					Text string `json:"text"`
//				} `json:"parts"`
//				Role string `json:"role"`
//			} `json:"content"`
//			FinishReason     string `json:"finishReason"`
//			CitationMetadata struct {
//				CitationSources []struct {
//					StartIndex int    `json:"startIndex"`
//					EndIndex   int    `json:"endIndex"`
//					URI        string `json:"uri,omitempty"`
//				} `json:"citationSources"`
//			} `json:"citationMetadata"`
//			AvgLogprobs float64 `json:"avgLogprobs"`
//		} `json:"candidates"`
//		UsageMetadata struct {
//			PromptTokenCount     int `json:"promptTokenCount"`
//			CandidatesTokenCount int `json:"candidatesTokenCount"`
//			TotalTokenCount      int `json:"totalTokenCount"`
//			PromptTokensDetails  []struct {
//				Modality   string `json:"modality"`
//				TokenCount int    `json:"tokenCount"`
//			} `json:"promptTokensDetails"`
//			CandidatesTokensDetails []struct {
//				Modality   string `json:"modality"`
//				TokenCount int    `json:"tokenCount"`
//			} `json:"candidatesTokensDetails"`
//		} `json:"usageMetadata"`
//		ModelVersion string `json:"modelVersion"`
//	}
//
// But only necessary part is utilizing
type GeminiOriginResponse struct {
	Candidates   []Candidate `json:"candidates"`
	ModelVersion string      `json:"modelVersion"`
}

type Candidate struct {
	Content `json:"content"`
}

type Content struct {
	Parts []Part `json:"parts"`
	Role  string `json:"role"`
}

type Part []struct {
	Text string `json:"text"`
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
func (r *GeminiProxy) Send(request GeminiProxyRequest) (*GeminiProxyResponse, error) {
	res := &GeminiProxyResponse{}
	geminiApiKey := os.Getenv("GEMINI_API_KEY")
	if geminiApiKey == "" {
		return nil, fmt.Errorf("gemini API key is not found")
	}

	httpReq, err := http.NewRequest("POST", r.OriginURL+geminiApiKey, nil)
	if err != nil {
		log.Printf("[ERROR] cannot create POST request: %#v;", err)
		return nil, err
	}

	httpResp, err := r.Client.Do(httpReq)
	if err != nil {
		log.Printf("[ERROR] can not make Get request: %#v", err)
		if errClose := httpResp.Body.Close(); errClose != nil {
			log.Printf("[ERROR] can not close response body %#v", errClose)
		}
		return nil, err
	}

	if httpResp == nil {
		return nil, fmt.Errorf("response from Gemini is nil")
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response from Gemini is nil or the response is not 200: %s", httpResp.Status)
	}

	byteResp, err := io.ReadAll(httpResp.Body)
	if err != nil {
		log.Printf("[ERROR] can not read response body %#v", err)
		return nil, err
	}

	err = json.Unmarshal(byteResp, &res)
	if err != nil {
		log.Printf("[ERROR] cannot unmarshal response %#v", err)
		return nil, err
	}

	return res, nil
}

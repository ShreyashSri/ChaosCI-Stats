package naas

import (
	"encoding/json"
	"net/http"
	"time"
)

type NaaSResponse struct {
	Reason string `json:"reason"`
}

var client = &http.Client{
	Timeout: 2 * time.Second,
}

// GetReason calls the No-as-a-Service API and returns a rejection reason.
// If the request fails or times out, it returns a fallback reason.
func GetReason() string {
	fallback := "Computer says no."

	req, err := http.NewRequest(http.MethodGet, "https://naas.isalman.dev/no", nil)
	if err != nil {
		return fallback
	}

	resp, err := client.Do(req)
	if err != nil {
		return fallback
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fallback
	}

	var data NaaSResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return fallback
	}

	if data.Reason == "" {
		return fallback
	}

	return data.Reason
}

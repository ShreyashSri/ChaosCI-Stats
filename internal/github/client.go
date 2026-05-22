package github

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Client interface {
	CreateCheckRun(ctx context.Context, repo string, commitSha string) (int64, error)
	UpdateCheckRun(ctx context.Context, repo string, checkID int64, status string, conclusion string, output string) error
	GetFileContent(ctx context.Context, repo string, commitSha string, path string) ([]byte, error)
}

type appClient struct {
	appID      string
	privateKey *rsa.PrivateKey
	httpClient *http.Client
}

func NewClient() (Client, error) {
	appID := os.Getenv("GITHUB_APP_ID")
	if appID == "" {
		return nil, fmt.Errorf("GITHUB_APP_ID is not set")
	}

	keyBytes, err := os.ReadFile(os.Getenv("GITHUB_APP_PRIVATE_KEY_PATH"))
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %v", err)
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	return &appClient{
		appID:      appID,
		privateKey: privateKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *appClient) generateJWT() (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now.Add(-60 * time.Second)),
		ExpiresAt: jwt.NewNumericDate(now.Add(10 * time.Minute)),
		Issuer:    c.appID,
	})
	return token.SignedString(c.privateKey)
}

func (c *appClient) getInstallationToken(ctx context.Context, repo string) (string, error) {
	jwtToken, err := c.generateJWT()
	if err != nil {
		return "", err
	}

	// 1. Get installation ID for the repo
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/repos/"+repo+"/installation", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get installation for repo %s: %s", repo, string(body))
	}

	var instRes struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&instRes); err != nil {
		return "", err
	}

	// 2. Create installation access token
	req, err = http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", instRes.ID), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to create access token: %s", string(body))
	}

	var tokenRes struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenRes); err != nil {
		return "", err
	}

	return tokenRes.Token, nil
}

func (c *appClient) CreateCheckRun(ctx context.Context, repo string, commitSha string) (int64, error) {
	token, err := c.getInstallationToken(ctx, repo)
	if err != nil {
		return 0, err
	}

	payload := map[string]interface{}{
		"name":       "ChaosCI Stats",
		"head_sha":   commitSha,
		"status":     "in_progress",
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.github.com/repos/"+repo+"/check-runs", bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("failed to create check run: %s", string(b))
	}

	var res struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return 0, err
	}
	return res.ID, nil
}

func (c *appClient) UpdateCheckRun(ctx context.Context, repo string, checkID int64, status string, conclusion string, output string) error {
	token, err := c.getInstallationToken(ctx, repo)
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"status": status,
	}
	if conclusion != "" {
		payload["conclusion"] = conclusion
	}
	if output != "" {
		payload["output"] = map[string]string{
			"title":   "ChaosCI Result",
			"summary": output,
		}
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "PATCH", fmt.Sprintf("https://api.github.com/repos/%s/check-runs/%d", repo, checkID), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update check run: %s", string(b))
	}
	return nil
}

func (c *appClient) GetFileContent(ctx context.Context, repo string, commitSha string, path string) ([]byte, error) {
	token, err := c.getInstallationToken(ctx, repo)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.github.com/repos/%s/contents/%s?ref=%s", repo, path, commitSha), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3.raw")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch file %s: %s", path, string(b))
	}

	return io.ReadAll(resp.Body)
}

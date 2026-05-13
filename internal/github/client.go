package github

import (
	"context"
	"log"
)

type Client interface {
	CreateCheckRun(ctx context.Context, repo string, commitSha string) (int64, error)
	UpdateCheckRun(ctx context.Context, repo string, checkID int64, status string, conclusion string, output string) error
}

type MockClient struct{}

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (m *MockClient) CreateCheckRun(ctx context.Context, repo string, commitSha string) (int64, error) {
	log.Printf("GitHub CheckRun created for %s@%s", repo, commitSha)
	return 12345, nil
}

func (m *MockClient) UpdateCheckRun(ctx context.Context, repo string, checkID int64, status string, conclusion string, output string) error {
	log.Printf("GitHub CheckRun %d updated: status=%s, conclusion=%s", checkID, status, conclusion)
	return nil
}

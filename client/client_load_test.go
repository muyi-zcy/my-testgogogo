package client

import (
	"testing"
	"time"
)

func TestCloneSharesLoadTransport(t *testing.T) {
	t.Parallel()

	base := NewForLoadWithRouter("http://localhost:8080", time.Second, nil)
	clone := base.Clone()
	if base.httpClient.Transport != clone.httpClient.Transport {
		t.Fatal("Clone should share Transport for connection reuse")
	}
	if clone.authToken != base.authToken {
		t.Fatal("Clone should copy auth token")
	}
}

package client

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type stubRouter struct {
	resolve func(path string) string
}

func (s stubRouter) Resolve(path string) string {
	return s.resolve(path)
}

func TestBuildURLWithRouter(t *testing.T) {
	t.Parallel()

	c := NewWithRouter("http://default:8080", time.Second, stubRouter{
		resolve: func(path string) string {
			if path == "/wcs/device/read" || path == "wcs/device/read" {
				return "http://wcs:20003"
			}
			return ""
		},
	})

	fullURL, err := c.buildURL("/wcs/device/read", url.Values{"id": {"1"}})
	require.NoError(t, err)
	require.Equal(t, "http://wcs:20003/wcs/device/read?id=1", fullURL)

	defaultURL, err := c.buildURL("/wms/in/create", nil)
	require.NoError(t, err)
	require.Equal(t, "http://default:8080/wms/in/create", defaultURL)
}

func TestBuildURLWithoutRouter(t *testing.T) {
	t.Parallel()

	c := New("http://default:8080", time.Second)
	fullURL, err := c.buildURL("/wms/in/create", nil)
	require.NoError(t, err)
	require.Equal(t, "http://default:8080/wms/in/create", fullURL)
}

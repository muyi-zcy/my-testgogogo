package config

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestRouterResolvePrefixLongestMatch(t *testing.T) {
	t.Parallel()

	router, err := buildRouter("http://default:8080", nil, RoutesYAML{
		Rules: []RouteRule{
			{Type: "prefix", Pattern: "/wms", Target: "http://wms:20002"},
			{Type: "prefix", Pattern: "/wms/sys", Target: "http://wms-auth:20002"},
		},
	})
	require.NoError(t, err)

	require.Equal(t, "http://wms-auth:20002", router.Resolve("/wms/sys/login"))
	require.Equal(t, "http://wms:20002", router.Resolve("/wms/in/create"))
	require.Equal(t, "http://default:8080", router.Resolve("/wmservice/list"))
	require.Equal(t, "http://default:8080", router.Resolve("/other"))
}

func TestRouterResolveRegex(t *testing.T) {
	t.Parallel()

	router, err := buildRouter("http://default:8080", nil, RoutesYAML{
		Rules: []RouteRule{
			{Type: "regex", Pattern: `^/wcs/`, Target: "http://wcs:20003"},
			{Type: "regex", Pattern: `^/wcs/device/`, Target: "http://wcs-device:20004"},
		},
	})
	require.NoError(t, err)

	require.Equal(t, "http://wcs:20003", router.Resolve("/wcs/device/read"))
	require.Equal(t, "http://default:8080", router.Resolve("/wms/in/create"))
}

func TestRouterResolvePrefixBeforeRegex(t *testing.T) {
	t.Parallel()

	router, err := buildRouter("http://default:8080", nil, RoutesYAML{
		Rules: []RouteRule{
			{Type: "prefix", Pattern: "/wms", Target: "http://wms:20002"},
			{Type: "regex", Pattern: `^/wms/special/`, Target: "http://wms-special:20005"},
		},
	})
	require.NoError(t, err)

	require.Equal(t, "http://wms:20002", router.Resolve("/wms/special/foo"))
}

func TestRoutesYAMLUnmarshalMap(t *testing.T) {
	t.Parallel()

	var env EnvConfig
	err := yaml.Unmarshal([]byte(`
base_url: http://default:8080
routes:
  /wms: http://wms:20002
  "regex:^/wcs/": http://wcs:20003
`), &env)
	require.NoError(t, err)
	require.Len(t, env.Routes.Rules, 2)
	require.Equal(t, "prefix", env.Routes.Rules[0].Type)
	require.Equal(t, "/wms", env.Routes.Rules[0].Pattern)
	require.Equal(t, "http://wms:20002", env.Routes.Rules[0].Target)
	require.Equal(t, "regex", env.Routes.Rules[1].Type)
	require.Equal(t, `^/wcs/`, env.Routes.Rules[1].Pattern)
}

func TestRoutesYAMLUnmarshalList(t *testing.T) {
	t.Parallel()

	var env EnvConfig
	err := yaml.Unmarshal([]byte(`
base_url: http://default:8080
services:
  wms: http://wms:20002
routes:
  - prefix: /wms
    target: wms
  - regex: '^/wcs/'
    target: http://wcs:20003
`), &env)
	require.NoError(t, err)

	router, err := buildRouter(env.BaseURL, env.Services, env.Routes)
	require.NoError(t, err)
	require.Equal(t, "http://wms:20002", router.Resolve("/wms/sys/login"))
	require.Equal(t, "http://wcs:20003", router.Resolve("/wcs/device/read"))
}

func TestBuildRouterInvalidRegex(t *testing.T) {
	t.Parallel()

	_, err := buildRouter("http://default:8080", nil, RoutesYAML{
		Rules: []RouteRule{{Type: "regex", Pattern: "[", Target: "http://bad:1"}},
	})
	require.Error(t, err)
}

func TestConfigCacheKeyWithRoutes(t *testing.T) {
	t.Parallel()

	router, err := buildRouter("http://default:8080", nil, RoutesYAML{
		Rules: []RouteRule{{Type: "prefix", Pattern: "/wms", Target: "http://wms:20002"}},
	})
	require.NoError(t, err)

	cfg := &Config{
		Active:  "local",
		BaseURL: "http://default:8080",
		Router:  router,
		User:    UserConfig{Username: "test"},
	}

	key := cfg.CacheKey()
	require.Contains(t, key, "p:/wms=http://wms:20002;")
}

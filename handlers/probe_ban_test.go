/*
Chronokeep Remote - Remote Tag Read API
Copyright (C) 2026 James Sentinella

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package handlers

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestIsProbePath(t *testing.T) {
	probes := []string{
		"/.env", "/.env.production", "/api/.env", "/.git/config", "/.aws/credentials",
		"/index.php", "/wp-login.php", "/admin/config.PHP", "/shell.aspx", "/cgi-bin/luci",
		"/wp-admin/", "/wp-content/plugins/x/readme.txt", "/phpmyadmin/index.html", "/pma",
		"/vendor/phpunit/phpunit/src/Util/PHP/eval-stdin.php", "/xmlrpc", "/.DS_Store",
		"/static/../../etc/passwd", "/.ssh/id_rsa", "/.htaccess",
	}
	for _, path := range probes {
		assert.True(t, IsProbePath(path), "expected %q to be a probe", path)
	}
	fine := []string{
		"/", "/health", "/results", "/results/stream", "/event-year/add", "/account/password",
		"/r/participants", "/environment", "/github", "/php-runners",
	}
	for _, path := range fine {
		assert.False(t, IsProbePath(path), "expected %q not to be a probe", path)
	}
}

func TestIsProbePathMatchesNoRoute(t *testing.T) {
	e := echo.New()
	h := Handler{}
	h.Bind(e.Group(""))
	h.BindRestricted(e.Group(""))
	routes := e.Router().Routes()
	assert.NotEmpty(t, routes)
	for _, route := range routes {
		assert.False(t, IsProbePath(route.Path), "route %s %s would get people banned", route.Method, route.Path)
	}
}

// probeServer runs a ProbeBanner in front of a route that always answers 204.
func probeServer(t *testing.T, allowlist ...string) (*ProbeBanner, func(ip, path string) int) {
	t.Helper()
	banner, err := NewProbeBanner(allowlist)
	if err != nil {
		t.Fatalf("Error making probe banner: %v", err)
	}
	e := echo.New()
	e.Pre(banner.Middleware())
	e.GET("/health", func(c *echo.Context) error { return c.NoContent(http.StatusNoContent) })
	request := func(ip, path string) int {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.RemoteAddr = net.JoinHostPort(ip, "40000")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		return rec.Code
	}
	return banner, request
}

func TestProbeBannerBansProbingIPs(t *testing.T) {
	_, request := probeServer(t)
	assert.Equal(t, http.StatusNoContent, request("203.0.113.5", "/health"))

	// the probe looks like nothing is there, and bans the IP
	assert.Equal(t, http.StatusNotFound, request("203.0.113.5", "/.env"))
	assert.Equal(t, http.StatusForbidden, request("203.0.113.5", "/health"))
	assert.Equal(t, http.StatusForbidden, request("203.0.113.5", "/results"))

	// other IPs aren't affected, including IPv6
	assert.Equal(t, http.StatusNoContent, request("198.51.100.7", "/health"))
	assert.Equal(t, http.StatusNotFound, request("2001:db8::1", "/wp-login.php"))
	assert.Equal(t, http.StatusForbidden, request("2001:db8::1", "/health"))
	assert.Equal(t, http.StatusNoContent, request("2001:db8::2", "/health"))
}

func TestProbeBanExpires(t *testing.T) {
	banner, request := probeServer(t)
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	banner.now = func() time.Time { return now }

	request("203.0.113.5", "/index.php")
	now = now.Add(ProbeBanDuration - time.Second)
	assert.Equal(t, http.StatusForbidden, request("203.0.113.5", "/health"))
	now = now.Add(time.Second)
	assert.Equal(t, http.StatusNoContent, request("203.0.113.5", "/health"))
}

func TestProbeBannerNeverBansLocalOrAllowedIPs(t *testing.T) {
	_, request := probeServer(t, "198.51.100.0/24", "2001:db8::5")
	for _, ip := range []string{"127.0.0.1", "::1", "10.1.2.3", "192.168.1.20", "172.16.0.9", "fd00::1", "198.51.100.40", "2001:db8::5"} {
		assert.Equal(t, http.StatusNotFound, request(ip, "/.env"), ip)
		assert.Equal(t, http.StatusNoContent, request(ip, "/health"), "%s shouldn't be banned", ip)
	}
	// just outside the allowed range is banned as usual
	request("198.51.101.1", "/.env")
	assert.Equal(t, http.StatusForbidden, request("198.51.101.1", "/health"))
}

func TestNewProbeBannerRejectsBadAllowlist(t *testing.T) {
	_, err := NewProbeBanner([]string{"203.0.113.0/24", "not-an-ip"})
	assert.Error(t, err)
	banner, err := NewProbeBanner([]string{"", " 203.0.113.9 "})
	if assert.NoError(t, err) {
		assert.Len(t, banner.allow, 1)
	}
}

func TestProbeBannerSweepsExpiredBans(t *testing.T) {
	banner, request := probeServer(t)
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	banner.now = func() time.Time { return now }

	request("203.0.113.5", "/.env")
	request("203.0.113.6", "/.env")
	now = now.Add(ProbeBanDuration + time.Minute)
	request("203.0.113.7", "/.env")
	banner.mu.Lock()
	defer banner.mu.Unlock()
	assert.Len(t, banner.bans, 1)
}

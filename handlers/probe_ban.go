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
	"fmt"
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v5"
	log "github.com/sirupsen/logrus"
)

// ProbeBanDuration is how long an IP stays banned after asking for a probe path.
const ProbeBanDuration = 24 * time.Hour

// probeSuffixes are file types this API never serves, which scanners ask for when looking
// for exploitable scripts.
var probeSuffixes = []string{
	".php", ".php5", ".php7", ".phtml",
	".asp", ".aspx", ".jsp", ".cgi",
}

// probeSegments are path segments (or the start of one) for secrets, version control and
// admin tools that scanners look for.
var probeSegments = []string{
	".env", ".git", ".svn", ".hg", ".aws", ".ssh", ".htaccess", ".htpasswd", ".ds_store",
	"wp-admin", "wp-login", "wp-content", "wp-includes", "xmlrpc",
	"phpmyadmin", "pma", "cgi-bin", "phpunit",
}

// IsProbePath reports whether a request path is one only a scanner looking for exploits
// would ask for. Nothing this API serves matches.
func IsProbePath(path string) bool {
	lower := strings.ToLower(path)
	if strings.Contains(lower, "../") || strings.Contains(lower, "..\\") {
		return true
	}
	for _, suffix := range probeSuffixes {
		if strings.HasSuffix(lower, suffix) {
			return true
		}
	}
	for _, segment := range strings.Split(lower, "/") {
		for _, probe := range probeSegments {
			if segment == probe || (strings.HasPrefix(probe, ".") && strings.HasPrefix(segment, probe)) {
				return true
			}
		}
	}
	return false
}

// ProbeBanner bans IPs that ask for probe paths. Bans are kept in memory, so they only last
// until the server restarts. Loopback and private addresses are never banned, so a proxy
// or someone on the same network can't be locked out, and neither is anything in the allowlist.
type ProbeBanner struct {
	mu        sync.Mutex
	bans      map[netip.Addr]time.Time
	allow     []netip.Prefix
	duration  time.Duration
	now       func() time.Time
	lastSweep time.Time
}

// NewProbeBanner makes a ProbeBanner. allowlist holds IPs or CIDRs that are never banned.
func NewProbeBanner(allowlist []string) (*ProbeBanner, error) {
	b := &ProbeBanner{
		bans:     make(map[netip.Addr]time.Time),
		duration: ProbeBanDuration,
		now:      time.Now,
	}
	for _, entry := range allowlist {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if prefix, err := netip.ParsePrefix(entry); err == nil {
			b.allow = append(b.allow, prefix.Masked())
			continue
		}
		addr, err := netip.ParseAddr(entry)
		if err != nil {
			return nil, fmt.Errorf("invalid ban allowlist entry %q: not an IP or CIDR", entry)
		}
		b.allow = append(b.allow, netip.PrefixFrom(addr.Unmap(), addr.Unmap().BitLen()))
	}
	return b, nil
}

// exempt reports whether an address can never be banned.
func (b *ProbeBanner) exempt(addr netip.Addr) bool {
	if addr.IsLoopback() || addr.IsPrivate() || addr.IsLinkLocalUnicast() || addr.IsUnspecified() {
		return true
	}
	for _, prefix := range b.allow {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}

// Banned reports whether an address is banned right now.
func (b *ProbeBanner) Banned(addr netip.Addr) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	until, ok := b.bans[addr]
	if !ok {
		return false
	}
	if !b.now().Before(until) {
		delete(b.bans, addr)
		return false
	}
	return true
}

// ban bans an address for the ban duration. It returns false for exempt addresses.
func (b *ProbeBanner) ban(addr netip.Addr) bool {
	if b.exempt(addr) {
		return false
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	now := b.now()
	b.bans[addr] = now.Add(b.duration)
	// drop expired bans now and then, so the map doesn't grow without end
	if now.Sub(b.lastSweep) > time.Minute {
		b.lastSweep = now
		for banned, until := range b.bans {
			if !now.Before(until) {
				delete(b.bans, banned)
			}
		}
	}
	return true
}

// Middleware turns away banned IPs with a 403 and bans IPs that ask for probe paths,
// answering those with a 404. Use it with Echo's Pre so it runs before routing and logging.
func (b *ProbeBanner) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			addr, err := netip.ParseAddr(c.RealIP())
			if err != nil {
				return next(c)
			}
			addr = addr.Unmap()
			if b.Banned(addr) {
				return c.NoContent(http.StatusForbidden)
			}
			if IsProbePath(c.Request().URL.Path) {
				if b.ban(addr) {
					log.Warnf("Banned %s for %v after it asked for %q.", addr, b.duration, c.Request().URL.Path)
				}
				return c.NoContent(http.StatusNotFound)
			}
			return next(c)
		}
	}
}

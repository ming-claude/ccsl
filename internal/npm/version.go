package npm

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	gojson "github.com/goccy/go-json"
)

const (
	distTagsURL      = "https://registry.npmjs.org/-/package/@anthropic-ai/claude-code/dist-tags"
	cacheFileName    = "npm-version.json"
	defaultTimeoutMs = 3000
	defaultCacheTTL  = 3600 // 1 hour
)

// VersionInfo holds the latest version from npm and whether an update is available.
type VersionInfo struct {
	Latest     string `json:"latest"`
	HasUpdate  bool   `json:"has_update"`
	CurrentVer string `json:"current_ver"`
}

// VersionClient checks npm for the latest Claude Code version with disk caching.
type VersionClient struct {
	CacheDir  string
	CacheTTL  int // seconds
	TimeoutMs int
}

type cacheEnvelope struct {
	FetchedAt int64             `json:"fetched_at"`
	Tags      map[string]string `json:"tags"`
}

// Check returns the latest version info, comparing against currentVersion.
// Returns nil on any failure (graceful degradation).
func (c *VersionClient) Check(currentVersion string) *VersionInfo {
	if currentVersion == "" {
		return nil
	}

	tags := c.getTags()
	target := tags["latest"]
	if target == "" {
		return nil
	}

	return &VersionInfo{
		Latest:     target,
		HasUpdate:  compareSemver(target, currentVersion) > 0,
		CurrentVer: currentVersion,
	}
}

func (c *VersionClient) getTags() map[string]string {
	if cached, age := c.readCache(); cached != nil && age <= c.ttl() {
		return cached
	}

	tags, err := c.fetchFromNpm()
	if err != nil {
		if cached, _ := c.readCache(); cached != nil {
			return cached
		}
		return nil
	}

	_ = c.writeCache(tags)
	return tags
}

func (c *VersionClient) fetchFromNpm() (map[string]string, error) {
	client := &http.Client{Timeout: c.timeout()}
	resp, err := client.Get(distTagsURL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return nil, err
	}

	tags := make(map[string]string)
	if err := gojson.Unmarshal(body, &tags); err != nil {
		return nil, err
	}
	return tags, nil
}

func (c *VersionClient) ttl() time.Duration {
	ttl := c.CacheTTL
	if ttl <= 0 {
		ttl = defaultCacheTTL
	}
	return time.Duration(ttl) * time.Second
}

func (c *VersionClient) timeout() time.Duration {
	ms := c.TimeoutMs
	if ms <= 0 {
		ms = defaultTimeoutMs
	}
	return time.Duration(ms) * time.Millisecond
}

func (c *VersionClient) cachePath() string {
	return filepath.Join(c.CacheDir, cacheFileName)
}

func (c *VersionClient) readCache() (map[string]string, time.Duration) {
	data, err := os.ReadFile(c.cachePath())
	if err != nil {
		return nil, 0
	}
	var env cacheEnvelope
	if err := gojson.Unmarshal(data, &env); err != nil {
		return nil, 0
	}
	if len(env.Tags) == 0 {
		return nil, 0
	}
	age := time.Since(time.Unix(env.FetchedAt, 0))
	return env.Tags, age
}

func (c *VersionClient) writeCache(tags map[string]string) error {
	path := c.cachePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	env := cacheEnvelope{FetchedAt: time.Now().Unix(), Tags: tags}
	data, err := gojson.Marshal(env)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}

// compareSemver compares two semver strings (X.Y.Z).
// Returns -1, 0, or 1 like strings.Compare but for version semantics.
func compareSemver(a, b string) int {
	ap := parseSemver(a)
	bp := parseSemver(b)
	for i := 0; i < 3; i++ {
		if ap[i] < bp[i] {
			return -1
		}
		if ap[i] > bp[i] {
			return 1
		}
	}
	return 0
}

func parseSemver(v string) [3]int {
	var parts [3]int
	for i, s := range strings.SplitN(v, ".", 3) {
		if i >= 3 {
			break
		}
		parts[i], _ = strconv.Atoi(s)
	}
	return parts
}

// ReadChannel returns the auto-update channel configured in
// <home>/.claude/settings.json's autoUpdatesChannel field.
// Returns "latest" on any failure (missing file, invalid JSON,
// missing or empty field) for graceful fallback.
func ReadChannel(home string) string {
	const fallback = "latest"
	if home == "" {
		return fallback
	}

	data, err := os.ReadFile(filepath.Join(home, ".claude", "settings.json"))
	if err != nil {
		return fallback
	}

	var settings struct {
		AutoUpdatesChannel string `json:"autoUpdatesChannel"`
	}
	if err := gojson.Unmarshal(data, &settings); err != nil {
		return fallback
	}
	if settings.AutoUpdatesChannel == "" {
		return fallback
	}
	return settings.AutoUpdatesChannel
}

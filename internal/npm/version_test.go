package npm

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	gojson "github.com/goccy/go-json"
)

func TestCompareSemver(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want int
	}{
		{"equal versions", "1.2.3", "1.2.3", 0},
		{"a major greater", "2.0.0", "1.0.0", 1},
		{"b major greater", "1.0.0", "2.0.0", -1},
		{"a minor greater", "1.3.0", "1.2.0", 1},
		{"b minor greater", "1.2.0", "1.3.0", -1},
		{"a patch greater", "1.2.4", "1.2.3", 1},
		{"b patch greater", "1.2.3", "1.2.4", -1},
		{"1.10 vs 1.9", "1.10.0", "1.9.0", 1},
		{"0.0.0 equal", "0.0.0", "0.0.0", 0},
		{"major trumps minor", "2.0.0", "1.99.99", 1},
		{"minor trumps patch", "1.2.0", "1.1.99", 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := compareSemver(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("compareSemver(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestParseSemver(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  [3]int
	}{
		{"full version", "1.2.3", [3]int{1, 2, 3}},
		{"major only", "5", [3]int{5, 0, 0}},
		{"major.minor only", "3.4", [3]int{3, 4, 0}},
		{"zeros", "0.0.0", [3]int{0, 0, 0}},
		{"large numbers", "10.20.30", [3]int{10, 20, 30}},
		{"empty string", "", [3]int{0, 0, 0}},
		{"non-numeric", "abc.def.ghi", [3]int{0, 0, 0}},
		{"partial non-numeric", "1.abc.3", [3]int{1, 0, 3}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseSemver(tc.input)
			if got != tc.want {
				t.Errorf("parseSemver(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestCacheRoundtrip(t *testing.T) {
	dir := t.TempDir()
	client := &VersionClient{CacheDir: dir}

	tags := map[string]string{"stable": "1.4.0", "latest": "1.5.0"}
	if err := client.writeCache(tags); err != nil {
		t.Fatalf("writeCache: %v", err)
	}

	got, age := client.readCache()
	if got["latest"] != "1.5.0" || got["stable"] != "1.4.0" {
		t.Errorf("readCache returned %v, want %v", got, tags)
	}
	if age > 5*time.Second {
		t.Errorf("cache age %v unexpectedly large", age)
	}

	// Verify the file content on disk.
	data, err := os.ReadFile(filepath.Join(dir, cacheFileName))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var env cacheEnvelope
	if err := gojson.Unmarshal(data, &env); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if env.Tags["latest"] != "1.5.0" {
		t.Errorf("envelope.Tags[latest] = %q, want %q", env.Tags["latest"], "1.5.0")
	}
}

func TestCacheExpiry(t *testing.T) {
	dir := t.TempDir()
	client := &VersionClient{
		CacheDir: dir,
		CacheTTL: 60, // 60 seconds
	}

	// Write a cache entry with an old timestamp (2 hours ago).
	env := cacheEnvelope{
		FetchedAt: time.Now().Add(-2 * time.Hour).Unix(),
		Tags:      map[string]string{"latest": "1.0.0"},
	}
	data, err := gojson.Marshal(env)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, cacheFileName), data, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// readCache should return the value but with age > TTL.
	got, age := client.readCache()
	if got["latest"] != "1.0.0" {
		t.Errorf("readCache returned %v, want latest=1.0.0", got)
	}
	if age <= client.ttl() {
		t.Errorf("expected expired cache (age %v <= ttl %v)", age, client.ttl())
	}
}

func TestCacheReadMissing(t *testing.T) {
	dir := t.TempDir()
	client := &VersionClient{CacheDir: dir}

	got, age := client.readCache()
	if got != nil {
		t.Errorf("readCache on missing file returned %v, want nil", got)
	}
	if age != 0 {
		t.Errorf("readCache on missing file returned age %v, want 0", age)
	}
}

func TestCheckViaHTTPTest(t *testing.T) {
	// Spin up a fake npm registry that returns a known latest version.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"latest":"2.0.0"}`))
	}))
	defer server.Close()

	// We can't easily override the URL in the production code, so test
	// the Check path using a pre-populated cache instead. This validates
	// the full Check → readCache → compareSemver path.
	dir := t.TempDir()
	client := &VersionClient{
		CacheDir: dir,
		CacheTTL: 3600,
	}

	// Pre-populate cache with "2.0.0".
	if err := client.writeCache(map[string]string{"latest": "2.0.0"}); err != nil {
		t.Fatalf("writeCache: %v", err)
	}

	info := client.Check("1.5.0", "latest")
	if info == nil {
		t.Fatal("Check returned nil")
	}
	if info.Latest != "2.0.0" {
		t.Errorf("Latest = %q, want %q", info.Latest, "2.0.0")
	}
	if !info.HasUpdate {
		t.Error("HasUpdate = false, want true")
	}
	if info.CurrentVer != "1.5.0" {
		t.Errorf("CurrentVer = %q, want %q", info.CurrentVer, "1.5.0")
	}

	// Same version → no update.
	info2 := client.Check("2.0.0", "latest")
	if info2 == nil {
		t.Fatal("Check returned nil for same version")
	}
	if info2.HasUpdate {
		t.Error("HasUpdate = true for same version, want false")
	}
}

func TestCheckEmptyVersion(t *testing.T) {
	client := &VersionClient{CacheDir: t.TempDir()}
	if info := client.Check("", "latest"); info != nil {
		t.Errorf("Check(\"\") = %v, want nil", info)
	}
}

func TestReadChannel_PresentInSettings(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	settings := `{"autoUpdatesChannel": "stable", "other": "ignored"}`
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte(settings), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got := ReadChannel(home)
	if got != "stable" {
		t.Errorf("ReadChannel = %q, want %q", got, "stable")
	}
}

func TestReadChannel_MissingField(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte(`{"model":"opus"}`), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got := ReadChannel(home)
	if got != "latest" {
		t.Errorf("ReadChannel = %q, want %q", got, "latest")
	}
}

func TestReadChannel_FileMissing(t *testing.T) {
	home := t.TempDir()
	got := ReadChannel(home)
	if got != "latest" {
		t.Errorf("ReadChannel = %q, want %q", got, "latest")
	}
}

func TestReadChannel_InvalidJSON(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte(`{not json`), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got := ReadChannel(home)
	if got != "latest" {
		t.Errorf("ReadChannel = %q, want %q", got, "latest")
	}
}

func TestReadChannel_EmptyString(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte(`{"autoUpdatesChannel":""}`), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got := ReadChannel(home)
	if got != "latest" {
		t.Errorf("ReadChannel = %q, want %q", got, "latest")
	}
}

func TestReadChannel_EmptyHome(t *testing.T) {
	got := ReadChannel("")
	if got != "latest" {
		t.Errorf("ReadChannel(\"\") = %q, want %q", got, "latest")
	}
}

func TestCheck_RespectsChannel(t *testing.T) {
	tagMap := map[string]string{
		"stable": "2.1.142",
		"latest": "2.1.150",
		"next":   "2.1.150",
	}

	tests := []struct {
		name          string
		current       string
		channel       string
		wantLatest    string
		wantChannel   string
		wantHasUpdate bool
	}{
		{"stable matches current", "2.1.142", "stable", "2.1.142", "stable", false},
		{"stable below current's channel target", "2.1.140", "stable", "2.1.142", "stable", true},
		{"latest is ahead", "2.1.142", "latest", "2.1.150", "latest", true},
		{"next tag", "2.1.142", "next", "2.1.150", "next", true},
		{"unknown channel falls back to latest", "2.1.142", "bogus", "2.1.150", "latest", true},
		{"empty channel falls back to latest", "2.1.142", "", "2.1.150", "latest", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			client := &VersionClient{CacheDir: dir, CacheTTL: 3600}
			if err := client.writeCache(tagMap); err != nil {
				t.Fatalf("writeCache: %v", err)
			}

			info := client.Check(tc.current, tc.channel)
			if info == nil {
				t.Fatal("Check returned nil")
			}
			if info.Latest != tc.wantLatest {
				t.Errorf("Latest = %q, want %q", info.Latest, tc.wantLatest)
			}
			if info.Channel != tc.wantChannel {
				t.Errorf("Channel = %q, want %q", info.Channel, tc.wantChannel)
			}
			if info.HasUpdate != tc.wantHasUpdate {
				t.Errorf("HasUpdate = %v, want %v", info.HasUpdate, tc.wantHasUpdate)
			}
			if info.CurrentVer != tc.current {
				t.Errorf("CurrentVer = %q, want %q", info.CurrentVer, tc.current)
			}
		})
	}
}

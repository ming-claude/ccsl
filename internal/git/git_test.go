package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// initGitRepo creates a bare git repo (git init) in the given directory.
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "init", dir)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}
	// Set user identity for commits.
	for _, args := range [][]string{
		{"config", "user.email", "test@test.com"},
		{"config", "user.name", "Test"},
	} {
		cmd = exec.Command("git", args...)
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			t.Fatalf("git config: %v", err)
		}
	}
}

// commitFile creates a file and commits it in the given repo.
func commitFile(t *testing.T, dir, name, content, msg string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	for _, args := range [][]string{
		{"add", name},
		{"commit", "-m", msg},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=Test", "GIT_COMMITTER_NAME=Test",
			"GIT_AUTHOR_EMAIL=test@test.com", "GIT_COMMITTER_EMAIL=test@test.com")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}

func TestGetData_GitRepo(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	commitFile(t, dir, "hello.txt", "hello world\n", "initial commit")

	data, err := GetData(dir, nil)
	if err != nil {
		t.Fatalf("GetData: %v", err)
	}
	if data == nil {
		t.Fatal("expected non-nil GitData for a git repo")
	}

	// Branch should be "main" or "master" depending on git defaults.
	if data.Branch == "" || data.Branch == "(init)" {
		t.Errorf("expected a real branch name, got %q", data.Branch)
	}

	// Clean repo with one commit.
	if !data.IsClean {
		t.Error("expected IsClean=true for committed repo")
	}
}

func TestGetData_NonGitDir(t *testing.T) {
	dir := t.TempDir()

	data, err := GetData(dir, nil)
	if err != nil {
		t.Fatalf("GetData: %v", err)
	}
	if data != nil {
		t.Fatalf("expected nil for non-git dir, got %+v", data)
	}
}

func TestGetData_UnbornHead(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	data, err := GetData(dir, nil)
	if err != nil {
		t.Fatalf("GetData: %v", err)
	}
	if data == nil {
		t.Fatal("expected non-nil GitData for init-only repo")
	}

	// On an unborn HEAD, branch --show-current returns the branch name
	// but rev-parse --short HEAD fails, so we may get the default branch
	// name or "(init)" depending on git version. Accept both.
	if data.Branch == "" {
		t.Error("expected non-empty branch for init-only repo")
	}
}

func TestParseStatusV2(t *testing.T) {
	tests := []struct {
		name string
		out  string
		want GitData
	}{
		{
			name: "clean repo with upstream",
			out:  "# branch.oid abc123\n# branch.head main\n# branch.upstream origin/main\n# branch.ab +0 -0",
			want: GitData{
				Branch:    "main",
				HasRemote: true,
				IsClean:   true,
			},
		},
		{
			name: "ahead and behind",
			out:  "# branch.oid abc123\n# branch.head feature\n# branch.upstream origin/feature\n# branch.ab +3 -2",
			want: GitData{
				Branch:    "feature",
				HasRemote: true,
				Ahead:     3,
				Behind:    2,
				IsClean:   true,
			},
		},
		{
			name: "modified and untracked files",
			out:  "# branch.oid abc123\n# branch.head main\n# branch.upstream origin/main\n# branch.ab +0 -0\n1 .M N... 100644 100644 100644 abc123 def456 file.go\n? untracked.txt",
			want: GitData{
				Branch:    "main",
				HasRemote: true,
				Modified:  1,
				Untracked: 1,
				IsClean:   false,
			},
		},
		{
			name: "added and deleted files",
			out:  "# branch.oid abc123\n# branch.head main\n1 A. N... 000000 100644 100644 000000 abc123 new.go\n1 .D N... 100644 000000 000000 abc123 000000 old.go",
			want: GitData{
				Branch:  "main",
				Added:   1,
				Deleted: 1,
				IsClean: false,
			},
		},
		{
			name: "rename entry",
			out:  "# branch.oid abc123\n# branch.head main\n2 R. N... 100644 100644 100644 abc123 def456 R100 new.go\told.go",
			want: GitData{
				Branch:   "main",
				Modified: 1,
				IsClean:  false,
			},
		},
		{
			name: "unmerged entry",
			out:  "# branch.oid abc123\n# branch.head main\nu UU N... 100644 100644 100644 100644 abc123 def456 789012 conflict.go",
			want: GitData{
				Branch:   "main",
				Modified: 1,
				IsClean:  false,
			},
		},
		{
			name: "multiple untracked",
			out:  "# branch.oid abc123\n# branch.head dev\n? a.txt\n? b.txt\n? c.txt",
			want: GitData{
				Branch:    "dev",
				Untracked: 3,
				IsClean:   false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got GitData
			parseStatusV2(tt.out, &got)
			if got != tt.want {
				t.Errorf("parseStatusV2() =\n  %+v\nwant\n  %+v", got, tt.want)
			}
		})
	}
}

func TestParseStatusV2_DetachedHead(t *testing.T) {
	out := "# branch.oid abc123def456\n# branch.head (detached)"
	var got GitData
	parseStatusV2(out, &got)

	if got.Branch != "(detached)" {
		t.Errorf("Branch = %q, want %q", got.Branch, "(detached)")
	}
	if got.HasRemote {
		t.Error("expected HasRemote=false for detached HEAD")
	}
	if !got.IsClean {
		t.Error("expected IsClean=true for detached HEAD with no changes")
	}
}

func TestParseStatusV2_NoUpstream(t *testing.T) {
	out := "# branch.oid abc123\n# branch.head feature-local"
	var got GitData
	parseStatusV2(out, &got)

	if got.Branch != "feature-local" {
		t.Errorf("Branch = %q, want %q", got.Branch, "feature-local")
	}
	if got.HasRemote {
		t.Error("expected HasRemote=false when no branch.ab line")
	}
	if got.Ahead != 0 || got.Behind != 0 {
		t.Errorf("expected Ahead=0, Behind=0, got Ahead=%d, Behind=%d", got.Ahead, got.Behind)
	}
	if !got.IsClean {
		t.Error("expected IsClean=true")
	}
}

func TestParseStatusV2_Empty(t *testing.T) {
	var got GitData
	parseStatusV2("", &got)

	if got.Branch != "(init)" {
		t.Errorf("Branch = %q, want %q", got.Branch, "(init)")
	}
	if !got.IsClean {
		t.Error("expected IsClean=true for empty output")
	}
}

func TestGetData_DirtyRepo(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	commitFile(t, dir, "base.txt", "base\n", "initial")

	// Create an untracked file.
	if err := os.WriteFile(filepath.Join(dir, "untracked.txt"), []byte("new\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Modify tracked file.
	if err := os.WriteFile(filepath.Join(dir, "base.txt"), []byte("changed\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	data, err := GetData(dir, nil)
	if err != nil {
		t.Fatalf("GetData: %v", err)
	}
	if data == nil {
		t.Fatal("expected non-nil GitData")
	}

	if data.IsClean {
		t.Error("expected IsClean=false for dirty repo")
	}
	if data.Untracked < 1 {
		t.Errorf("expected Untracked >= 1, got %d", data.Untracked)
	}
	if data.Modified < 1 {
		t.Errorf("expected Modified >= 1, got %d", data.Modified)
	}
}

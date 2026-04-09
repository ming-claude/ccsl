package segment

import (
	"testing"

	"github.com/ming-claude/ccsl/internal/git"
)

func TestGitBranch(t *testing.T) {
	seg := &GitBranchSegment{}
	ctx := &RenderContext{
		Git: &git.GitData{
			Branch:    "main",
			HasRemote: true,
			Ahead:     2,
			Behind:    0,
		},
	}
	out, err := seg.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Primary != "main" {
		t.Errorf("Primary: got %q, want %q", out.Primary, "main")
	}
	if out.Detail != "↑2" {
		t.Errorf("Detail: got %q, want %q", out.Detail, "↑2")
	}
}

func TestGitBranch_NilGit(t *testing.T) {
	seg := &GitBranchSegment{}
	ctx := &RenderContext{Git: nil}
	out, err := seg.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil output, got %+v", out)
	}
}

func TestGitChanges_Clean(t *testing.T) {
	seg := &GitChangesSegment{}
	ctx := &RenderContext{
		Git: &git.GitData{IsClean: true},
	}
	out, err := seg.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil output for clean repo, got %+v", out)
	}
}

func TestGitChanges_Dirty(t *testing.T) {
	seg := &GitChangesSegment{}
	ctx := &RenderContext{
		Git: &git.GitData{
			IsClean:  false,
			Modified: 3,
		},
	}
	out, err := seg.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Primary != "~3" {
		t.Errorf("got %q, want %q", out.Primary, "~3")
	}
}

package segment

import (
	"testing"

	"github.com/ming-claude/ccsl/internal/git"
	"github.com/ming-claude/ccsl/internal/input"
)

func TestGitGroup_FullRender(t *testing.T) {
	ctx := &RenderContext{
		Git: &git.GitData{
			Branch:     "main",
			HasRemote:  true,
			Ahead:      2,
			Added:      1,
			Modified:   3,
			Insertions: 42,
			Deletions:  7,
		},
		Style: StylePlain,
	}

	g := &GitGroup{}
	out, err := g.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}

	// Should contain branch with ahead info, changes, and insertions/deletions.
	expected := "main ↑2 +1 ~3 +42/-7"
	if out.Primary != expected {
		t.Errorf("expected %q, got %q", expected, out.Primary)
	}
}

func TestGitGroup_BranchOnly(t *testing.T) {
	ctx := &RenderContext{
		Git: &git.GitData{
			Branch:  "feature/foo",
			IsClean: true,
		},
		Style: StylePlain,
	}

	g := &GitGroup{}
	out, err := g.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}

	if out.Primary != "feature/foo" {
		t.Errorf("expected %q, got %q", "feature/foo", out.Primary)
	}
}

func TestGitGroup_AllNilData(t *testing.T) {
	ctx := &RenderContext{
		Style: StylePlain,
	}

	g := &GitGroup{}
	out, err := g.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil output, got %+v", out)
	}
}

func TestGitGroup_DisabledChildren(t *testing.T) {
	ctx := &RenderContext{
		Git: &git.GitData{
			Branch:     "main",
			Added:      1,
			Insertions: 10,
			Deletions:  5,
		},
		Style: StylePlain,
		Disabled: map[string]bool{
			"git.changes":    true,
			"git.insertions": true,
		},
	}

	g := &GitGroup{}
	out, err := g.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}

	// Only branch and deletions should remain.
	expected := "main -5"
	if out.Primary != expected {
		t.Errorf("expected %q, got %q", expected, out.Primary)
	}
}

func TestGitGroup_AllChildrenDisabled(t *testing.T) {
	ctx := &RenderContext{
		Git: &git.GitData{
			Branch:     "main",
			Added:      1,
			Insertions: 10,
			Deletions:  5,
		},
		Style: StylePlain,
		Disabled: map[string]bool{
			"git.branch":     true,
			"git.worktree":   true,
			"git.changes":    true,
			"git.insertions": true,
			"git.deletions":  true,
		},
	}

	g := &GitGroup{}
	out, err := g.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil output when all children disabled, got %+v", out)
	}
}

func TestGitGroup_ChildProducesNil(t *testing.T) {
	// Branch exists but no changes, no insertions, no deletions.
	ctx := &RenderContext{
		Git: &git.GitData{
			Branch:  "develop",
			IsClean: true,
		},
		Style: StylePlain,
	}

	g := &GitGroup{}
	out, err := g.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Primary != "develop" {
		t.Errorf("expected %q, got %q", "develop", out.Primary)
	}
}

func TestGitGroup_WithWorktree(t *testing.T) {
	ctx := &RenderContext{
		Git: &git.GitData{
			Branch: "main",
		},
		Stdin: &input.StdinData{
			Worktree: &input.WorktreeInfo{
				Name:   "feature-wt",
				Branch: "feature/bar",
			},
		},
		Style: StylePlain,
	}

	g := &GitGroup{}
	out, err := g.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}

	expected := "main feature-wt (feature/bar)"
	if out.Primary != expected {
		t.Errorf("expected %q, got %q", expected, out.Primary)
	}
}

func TestGitGroup_InsertionsOnly(t *testing.T) {
	ctx := &RenderContext{
		Git: &git.GitData{
			Branch:     "main",
			Insertions: 15,
		},
		Style: StylePlain,
		Disabled: map[string]bool{
			"git.deletions": true,
		},
	}

	g := &GitGroup{}
	out, err := g.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}

	expected := "main +15"
	if out.Primary != expected {
		t.Errorf("expected %q, got %q", expected, out.Primary)
	}
}

func TestGitGroup_Name(t *testing.T) {
	g := &GitGroup{}
	if g.Name() != "git" {
		t.Errorf("expected name %q, got %q", "git", g.Name())
	}
}

func TestGitGroup_DefaultPriority(t *testing.T) {
	g := &GitGroup{}
	if g.DefaultPriority() != 6 {
		t.Errorf("expected priority 6, got %d", g.DefaultPriority())
	}
}

package segment

import (
	"fmt"
)

// TodoProgressSegment displays the current in-progress todo item and completion count.
type TodoProgressSegment struct{}

func (s *TodoProgressSegment) Name() string         { return "todo_progress" }
func (s *TodoProgressSegment) DefaultTitle() string { return "" }
func (s *TodoProgressSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("todo_progress", m)
}
func (s *TodoProgressSegment) DefaultPriority() int { return 4 }

func (s *TodoProgressSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	td := ctx.Transcript
	if td == nil || len(td.Todos) == 0 {
		return nil, nil
	}

	// Find the current in-progress todo.
	var current string
	for _, todo := range td.Todos {
		if todo.Status == "in_progress" {
			current = todo.Subject
			break
		}
	}

	progress := fmt.Sprintf("(%d/%d)", td.TodoCompleted, len(td.Todos))
	if current != "" {
		return &SegmentOutput{Primary: current + " " + progress}, nil
	}

	return &SegmentOutput{Primary: progress}, nil
}

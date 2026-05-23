package segment

import "strings"

// ModelSegment displays the model display name from stdin, folding in the
// reasoning effort level when present.
// e.g. "Opus 4.7 (1M | xhigh)" or "Opus 4.7 (xhigh)" or "Opus 4.7 (1M)"
type ModelSegment struct{}

func (s *ModelSegment) Name() string         { return "model" }
func (s *ModelSegment) DefaultTitle() string { return "" }
func (s *ModelSegment) DefaultIcon(m StyleMode) string {
	return resolveIcon("model", m)
}
func (s *ModelSegment) DefaultPriority() int { return 10 }

func (s *ModelSegment) Render(ctx *RenderContext) (*SegmentOutput, error) {
	stdin := ctx.Stdin
	if stdin == nil || stdin.Model == nil {
		return nil, nil
	}
	name := stdin.Model.DisplayName
	if name == "" {
		name = stdin.Model.ID
	}
	var effort string
	if stdin.Effort != nil {
		effort = stdin.Effort.Level
	}
	return &SegmentOutput{Primary: decorateModelName(name, effort)}, nil
}

// decorateModelName folds the effort level into the model name. If the name
// ends with a parenthesized suffix (e.g. CC's "Opus 4.7 (1M context)"), the
// literal "context" token is dropped and the effort is appended inside the
// parens. Otherwise the effort is wrapped in a fresh pair of parens. When no
// effort is present the name is returned with only the "context" trim applied.
func decorateModelName(name, effort string) string {
	open := strings.LastIndex(name, "(")
	if open >= 0 && strings.HasSuffix(name, ")") {
		prefix := name[:open]
		inner := name[open+1 : len(name)-1]

		// Drop the literal "context" token, keep the rest: "1M context" -> "1M".
		fields := strings.Fields(inner)
		kept := fields[:0]
		for _, f := range fields {
			if f != "context" {
				kept = append(kept, f)
			}
		}
		inner = strings.Join(kept, " ")

		if effort != "" {
			if inner == "" {
				inner = effort
			} else {
				inner += " | " + effort
			}
		}
		if inner == "" {
			return strings.TrimRight(prefix, " ")
		}
		return prefix + "(" + inner + ")"
	}

	if effort != "" {
		return name + " (" + effort + ")"
	}
	return name
}

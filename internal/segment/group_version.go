package segment

// VersionGroup displays the CLI version with an update indicator.
type VersionGroup struct{}

func (g *VersionGroup) Name() string         { return "version" }
func (g *VersionGroup) DefaultTitle() string { return "" }
func (g *VersionGroup) DefaultIcon(m StyleMode) string {
	return resolveIcon("version", m)
}
func (g *VersionGroup) DefaultPriority() int { return 5 }

func (g *VersionGroup) Render(ctx *RenderContext) (*SegmentOutput, error) {
	vi := ctx.NpmVersion
	if vi == nil || !vi.HasUpdate {
		return nil, nil
	}
	stdin := ctx.Stdin
	ver := vi.Latest
	if stdin != nil && stdin.Version != "" {
		ver = stdin.Version + " \u2b06 " + vi.Latest
	}
	return &SegmentOutput{Primary: ver}, nil
}

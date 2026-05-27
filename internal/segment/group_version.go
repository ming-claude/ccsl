package segment

import "strings"

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
	channel := vi.Channel
	if channel == "" {
		channel = "latest"
	}
	ver := vi.Latest + " (" + channel + ")"
	if stdin != nil && stdin.Version != "" {
		target := stripCommonVersionPrefix(stdin.Version, vi.Latest)
		ver = stdin.Version + " \u2b06 " + target + " (" + channel + ")"
	}
	return &SegmentOutput{Primary: ver}, nil
}

// stripCommonVersionPrefix returns latest with leading dot-separated segments
// that match current removed. Compression only applies when the shared prefix
// covers at least major.minor (>= 2 segments) and one segment of latest still
// remains, so a major/minor bump is never displayed in a shape that looks like
// a patch bump.
func stripCommonVersionPrefix(current, latest string) string {
	if current == "" {
		return latest
	}
	cur := strings.Split(current, ".")
	lat := strings.Split(latest, ".")
	common := 0
	for i := 0; i < len(cur) && i < len(lat); i++ {
		if cur[i] != lat[i] {
			break
		}
		common++
	}
	if common < 2 || common >= len(lat) {
		return latest
	}
	return strings.Join(lat[common:], ".")
}

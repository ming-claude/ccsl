package style

import "github.com/ming-claude/ccsl/internal/segment"

func init() {
	segment.IconResolver = DefaultIcon
}

// Icon sets per group per style mode.
// key: group name, value: map[StyleMode]icon
var icons = map[string]map[segment.StyleMode]string{
	"model":        {segment.StylePlain: ">", segment.StyleNerdFont: "\U000F167A", segment.StylePowerline: "\U000F167A"},    // nf-md-head_cog
	"git":          {segment.StylePlain: "*", segment.StyleNerdFont: "\U000F02A2", segment.StylePowerline: "\U000F02A2"},    // nf-md-git
	"context":      {segment.StylePlain: "Ctx:", segment.StyleNerdFont: "\U000F035B", segment.StylePowerline: "\U000F035B"}, // nf-md-chip
	"tokens":       {segment.StylePlain: "Tok:", segment.StyleNerdFont: "\U000F0BCD", segment.StylePowerline: "\U000F0BCD"}, // nf-md-pound_box
	"cost":         {segment.StylePlain: "$", segment.StyleNerdFont: "\uef8d", segment.StylePowerline: "\uef8d"},            // nf-fa-sack_dollar
	"usage_5hour":  {segment.StylePlain: "%:", segment.StyleNerdFont: "\uf200", segment.StylePowerline: "\uf200"},           // nf-fa-chart_pie
	"usage_weekly": {segment.StylePlain: "%:", segment.StyleNerdFont: "\uf200", segment.StylePowerline: "\uf200"},           // nf-fa-chart_pie
	"session":      {segment.StylePlain: "S:", segment.StyleNerdFont: "\U000F0220", segment.StylePowerline: "\U000F0220"},   // nf-md-clock_outline
	"speed":        {segment.StylePlain: "Spd:", segment.StyleNerdFont: "\U000F04C5", segment.StylePowerline: "\U000F04C5"}, // nf-md-rocket
	"diff":         {segment.StylePlain: "+-:", segment.StyleNerdFont: "\U000F0992", segment.StylePowerline: "\U000F0992"},  // nf-md-compare
	"activity":     {segment.StylePlain: "~:", segment.StyleNerdFont: "\U000F0211", segment.StylePowerline: "\U000F0211"},   // nf-md-fire
	"live":         {segment.StylePlain: ">:", segment.StyleNerdFont: "\U000F0453", segment.StylePowerline: "\U000F0453"},   // nf-md-pulse
	"cwd":          {segment.StylePlain: "~/", segment.StyleNerdFont: "\U000F0770", segment.StylePowerline: "\U000F0770"},   // nf-md-folder_open
	"version":      {segment.StylePlain: "v:", segment.StyleNerdFont: "\uf409", segment.StylePowerline: "\uf409"},           // nf-oct-download
	"env":          {segment.StylePlain: "Env:", segment.StyleNerdFont: "\U000F0635", segment.StylePowerline: "\U000F0635"}, // nf-md-shield
	"clock":        {segment.StylePlain: "@", segment.StyleNerdFont: "\U000F0954", segment.StylePowerline: "\U000F0954"},    // nf-md-clock_outline
}

// DefaultIcon returns the default icon for a segment in the given style mode.
func DefaultIcon(segmentName string, mode segment.StyleMode) string {
	if m, ok := icons[segmentName]; ok {
		if icon, ok := m[mode]; ok {
			return icon
		}
	}
	return ""
}

var separators = map[segment.StyleMode]string{
	segment.StylePlain:     " | ",
	segment.StyleNerdFont:  " \u2502 ",
	segment.StylePowerline: "",
}

// DefaultSeparator returns the default separator for a style mode.
func DefaultSeparator(mode segment.StyleMode) string {
	if s, ok := separators[mode]; ok {
		return s
	}
	return " | "
}

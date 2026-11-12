package ui

import (
	glamouransi "charm.land/glamour/v2/ansi"
	"charm.land/glamour/v2/styles"
	"github.com/charmbracelet/colorprofile"
)

func (s *Styles) markdownStyle() glamouransi.StyleConfig {
	config := styles.NoTTYStyleConfig
	// Inline code consumes Prefix/Suffix, unlike block-level Markdown elements.
	config.Code.Prefix, config.Code.Suffix = "`", "`"
	if s.profile < colorprofile.ANSI {
		return config
	}
	yes := true
	for _, block := range []*glamouransi.StyleBlock{&config.Heading, &config.H1, &config.H2, &config.H3, &config.H4, &config.H5, &config.H6} {
		block.Bold = &yes
	}
	for _, primitive := range []*glamouransi.StylePrimitive{&config.Strong, &config.Emph, &config.Strikethrough, &config.Code.StylePrimitive} {
		primitive.BlockPrefix, primitive.BlockSuffix = "", ""
	}
	config.Code.Prefix, config.Code.Suffix = "", ""
	config.Strong.Bold, config.Emph.Italic, config.Strikethrough.CrossedOut = &yes, &yes, &yes
	config.LinkText.Underline, config.Link.Underline = &yes, &yes
	return config
}

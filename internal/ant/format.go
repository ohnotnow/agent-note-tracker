package ant

import (
	"fmt"
	"strings"
)

// RenderMarkdown emits the entries as a single markdown document. Each entry
// is wrapped in a YAML frontmatter block (id / kind / issue_id / created_at)
// followed by an H2 (the title, falling back to the public id) and the body.
// Multiple entries are separated by `***` (a horizontal rule that won't be
// confused with frontmatter delimiters).
func RenderMarkdown(entries []Entry) string {
	var sb strings.Builder
	for i, e := range entries {
		if i > 0 {
			sb.WriteString("\n***\n\n")
		}
		renderOne(&sb, e)
	}
	return sb.String()
}

func renderOne(sb *strings.Builder, e Entry) {
	sb.WriteString("---\n")
	fmt.Fprintf(sb, "id: %s\n", e.PublicID)
	fmt.Fprintf(sb, "kind: %s\n", e.Kind)
	if e.IssueID != "" {
		fmt.Fprintf(sb, "issue_id: %s\n", e.IssueID)
	}
	fmt.Fprintf(sb, "created_at: %s\n", e.CreatedAt)
	sb.WriteString("---\n\n")

	heading := e.Title
	if heading == "" {
		heading = e.PublicID
	}
	fmt.Fprintf(sb, "## %s\n\n", heading)
	sb.WriteString(e.Body)
	if !strings.HasSuffix(e.Body, "\n") {
		sb.WriteByte('\n')
	}
}

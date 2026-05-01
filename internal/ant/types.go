package ant

// Entry is the canonical record stored in the entries table.
//
// PublicID renders as "id" in JSON to match the user-facing identifier.
// The numeric primary key is intentionally never exposed.
type Entry struct {
	id        int64  // unexported: never serialised
	PublicID  string `json:"id"`
	Kind      string `json:"kind"`
	Title     string `json:"title,omitempty"`
	Body      string `json:"body"`
	IssueID   string `json:"issue_id,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Conventional kinds documented for the Claude Code skill. The schema does
// not enforce these — kind is a free string so projects can grow their own.
const (
	KindNote  = "note"
	KindADR   = "adr"
	KindPivot = "pivot"
)

// SnippetLen is the body-snippet length used by 'recent' and 'search'. The
// goal is "enough for the agent to decide whether to follow up with show",
// which 200 runes hits without bloating list output.
const SnippetLen = 200

// EntrySlim is the trimmed shape used as the default JSON output for list,
// recent, search, and for. Body is omitted on purpose; callers that want it
// pass --long (which returns the full Entry) or use show.
type EntrySlim struct {
	PublicID  string `json:"id"`
	Kind      string `json:"kind"`
	Title     string `json:"title,omitempty"`
	IssueID   string `json:"issue_id,omitempty"`
	CreatedAt string `json:"created_at"`
}

// EntryWithSnippet adds a SnippetLen-character body snippet to an EntrySlim.
// Used by 'recent' and 'search' so the agent can scan results without
// fetching every body in full.
type EntryWithSnippet struct {
	EntrySlim
	Snippet string `json:"snippet,omitempty"`
}

// Slim returns the slim projection of an Entry.
func (e Entry) Slim() EntrySlim {
	return EntrySlim{
		PublicID:  e.PublicID,
		Kind:      e.Kind,
		Title:     e.Title,
		IssueID:   e.IssueID,
		CreatedAt: e.CreatedAt,
	}
}

// WithSnippet returns the slim projection plus an n-rune body snippet.
func (e Entry) WithSnippet(n int) EntryWithSnippet {
	return EntryWithSnippet{
		EntrySlim: e.Slim(),
		Snippet:   makeSnippet(e.Body, n),
	}
}

func makeSnippet(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}

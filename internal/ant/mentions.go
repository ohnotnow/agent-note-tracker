package ant

import (
	"regexp"
	"strconv"
	"strings"
)

// EntryRef is another entry as shown alongside the one being read: its id and
// title, so a mistyped or lookalike id stands out next to the wrong title.
type EntryRef struct {
	PublicID string `json:"id"`
	Title    string `json:"title,omitempty"`
}

// EntryShown is the 'ant show' record: the entry plus the entries its body
// mentions and the entries whose bodies mention it. Mentions are worked out
// from the bodies on every read, so they never go stale.
//
// NotInNotebook lists ids with this project's prefix that match no entry.
// They are often ait issues (projects commonly share one prefix across ait
// and ant), possibly since closed, so they are listed without calling them
// typos.
type EntryShown struct {
	Entry
	Mentions      []EntryRef `json:"mentions,omitempty"`
	MentionedBy   []EntryRef `json:"mentioned_by,omitempty"`
	NotInNotebook []string   `json:"not_in_notebook,omitempty"`
}

// findMentions returns the distinct ids with the given prefix in body, in
// order of first appearance.
func findMentions(prefix, body string) []string {
	var (
		out  []string
		seen = map[string]bool{}
	)
	for _, span := range mentionSpans(prefix, body) {
		id := body[span[0]:span[1]]
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}

// rewriteMentions replaces each id with the given prefix in body that has an
// entry in renames. Ids missing from renames (ait issues sharing the prefix,
// say) are left alone.
func rewriteMentions(prefix, body string, renames map[string]string) string {
	var (
		b    strings.Builder
		last int
	)
	for _, span := range mentionSpans(prefix, body) {
		if to, ok := renames[body[span[0]:span[1]]]; ok {
			b.WriteString(body[last:span[0]])
			b.WriteString(to)
			last = span[1]
		}
	}
	b.WriteString(body[last:])
	return b.String()
}

// mentionSpans returns the [start, end) byte offsets of each id with the
// given prefix in body. An id must stand alone as a token, and one followed by
// a dotted number (an ait child issue such as x-AbCdE.3) is skipped.
func mentionSpans(prefix, body string) [][2]int {
	re := regexp.MustCompile(regexp.QuoteMeta(prefix) + `-[` + idAlphabet + `]{` + strconv.Itoa(idMinLength) + `,}(\.[0-9]+)*`)
	var out [][2]int
	for _, m := range re.FindAllStringSubmatchIndex(body, -1) {
		start, end := m[0], m[1]
		if m[2] >= 0 {
			continue // dotted suffix: an ait child issue
		}
		if start > 0 && isTokenByte(body[start-1]) {
			continue
		}
		if end < len(body) && isTokenByte(body[end]) {
			continue
		}
		out = append(out, [2]int{start, end})
	}
	return out
}

// isTokenByte reports whether b can be part of an id-like token, so an id
// glued to one of these is not a standalone mention.
func isTokenByte(b byte) bool {
	return b == '-' || b == '_' ||
		('0' <= b && b <= '9') || ('a' <= b && b <= 'z') || ('A' <= b && b <= 'Z')
}

// ShowEntry returns the entry with publicID plus its mentions. It reads every
// entry once to resolve titles and find mentioned_by; notebooks are hundreds
// of entries, so that is cheap.
func (s *Store) ShowEntry(prefix, publicID string) (EntryShown, error) {
	entry, err := s.GetEntry(publicID)
	if err != nil {
		return EntryShown{}, err
	}
	all, err := s.ListEntries(ListFilter{})
	if err != nil {
		return EntryShown{}, err
	}
	titles := make(map[string]string, len(all))
	for _, e := range all {
		titles[e.PublicID] = e.Title
	}

	shown := EntryShown{Entry: entry}
	for _, id := range findMentions(prefix, entry.Body) {
		if id == entry.PublicID {
			continue
		}
		if title, ok := titles[id]; ok {
			shown.Mentions = append(shown.Mentions, EntryRef{PublicID: id, Title: title})
		} else {
			shown.NotInNotebook = append(shown.NotInNotebook, id)
		}
	}
	for _, e := range all {
		if e.PublicID == entry.PublicID {
			continue
		}
		for _, id := range findMentions(prefix, e.Body) {
			if id == entry.PublicID {
				shown.MentionedBy = append(shown.MentionedBy, EntryRef{PublicID: e.PublicID, Title: e.Title})
				break
			}
		}
	}
	return shown, nil
}

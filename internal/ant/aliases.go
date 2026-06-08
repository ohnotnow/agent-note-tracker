package ant

import "strings"

// inputFlagAliases maps ait-compatibility flag spellings to their canonical
// ant flag names. ait and ant drifted on a couple of flag names; rather than
// reject the ait muscle-memory spelling, we accept it on input. These are
// INPUT-ONLY: they never appear in help text, usage strings, or output keys
// (the output key for kind stays "kind"). See epic ant-STIsz, section A.
var inputFlagAliases = map[string]string{
	"type":        "kind",
	"description": "body",
}

// canonicaliseAliases rewrites recognised input-flag aliases in args to their
// canonical ant names before flag parsing. Doing it up front means the rest of
// the pipeline (the FlagSet, extractPositional, setFlags) only ever sees
// canonical names, so help text and the "was this flag set" logic need no
// changes. It handles the --flag, -flag, and --flag=value forms, and leaves
// everything after a "--" terminator untouched.
//
// Supplying both an alias and its canonical name for the same field (e.g.
// --kind and --type together) is rejected, rather than silently letting the
// last one win — that ambiguity is almost always a mistake.
func canonicaliseAliases(args []string) ([]string, error) {
	out := make([]string, 0, len(args))
	suppliedAs := map[string]string{} // canonical name -> spelling first seen
	afterTerminator := false

	for _, tok := range args {
		if afterTerminator {
			out = append(out, tok)
			continue
		}
		if tok == "--" {
			afterTerminator = true
			out = append(out, tok)
			continue
		}

		name, rest, isFlag := splitFlagToken(tok)
		if !isFlag {
			out = append(out, tok)
			continue
		}

		canonical := name
		rewritten := tok
		if c, ok := inputFlagAliases[name]; ok {
			canonical = c
			rewritten = "--" + c + rest
		}

		if prev, seen := suppliedAs[canonical]; seen && prev != name {
			return nil, NewError(CodeUsage,
				"--%s and --%s set the same field; use one of them", prev, name)
		}
		suppliedAs[canonical] = name
		out = append(out, rewritten)
	}
	return out, nil
}

// splitFlagToken parses a CLI token into its flag name and the trailing
// "=value" remainder (if any). isFlag is false for non-flag tokens, the bare
// "-" (stdin marker), and "--" (terminator). Both -flag and --flag forms are
// recognised, mirroring Go's flag package.
func splitFlagToken(tok string) (name, rest string, isFlag bool) {
	if len(tok) < 2 || tok[0] != '-' {
		return "", "", false
	}
	s := tok[1:]
	if s[0] == '-' {
		s = s[1:]
	}
	if s == "" {
		return "", "", false
	}
	if i := strings.IndexByte(s, '='); i >= 0 {
		return s[:i], s[i:], true
	}
	return s, "", true
}

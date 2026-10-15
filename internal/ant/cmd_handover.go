package ant

import (
	"flag"
	"fmt"
)

// Handover handles 'ant handover' and 'ant handover create'.
//
// With no arguments it prints the newest handover entry in full, the same
// shape as 'ant show', and marks it shown so the session can append to it.
// "Newest" is by created_at, not by following supersedes links, so a chain
// that was linked wrongly by hand can't hide the current note.
//
// 'ant handover create' takes the same flags as 'ant add', with the kind set
// to handover and the previous handover (if any) superseded automatically.
func (a *App) Handover(args []string) error {
	if len(args) > 0 && args[0] == "create" {
		return a.handoverCreate(args[1:])
	}

	fs := flag.NewFlagSet("handover", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(a.Stderr, "usage: ant handover")
		fmt.Fprintln(a.Stderr, "  Print the newest handover entry in full.")
		fmt.Fprintln(a.Stderr, "usage: ant handover create <add flags>")
		fmt.Fprintln(a.Stderr, "  Add a handover entry that supersedes the previous one. Takes the")
		fmt.Fprintln(a.Stderr, "  same flags as 'ant add' except --kind and --supersedes.")
	}
	if err := a.parseFlags(fs, args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return NewError(CodeUsage, "usage: ant handover [create <add flags>]")
	}

	store, _, err := a.requireInitialised()
	if err != nil {
		return err
	}
	latest, err := latestHandover(store)
	if err != nil {
		return err
	}
	if latest == "" {
		return NewError(CodeNotFound, "no handover recorded, write one with: ant handover create --body @file.md")
	}
	prefix, _, err := store.GetPrefix()
	if err != nil {
		return err
	}
	entry, err := store.ShowEntry(prefix, latest)
	if err != nil {
		return err
	}
	if err := store.MarkShown(latest); err != nil {
		return err
	}
	return a.writeJSON(entry)
}

// handoverCreate runs 'ant add' with the kind and supersedes link filled in.
// --kind and --supersedes are refused rather than overridden, since either
// would quietly break the chain the command exists to keep.
func (a *App) handoverCreate(args []string) error {
	args, err := canonicaliseAliases(args)
	if err != nil {
		return err
	}
	for _, tok := range args {
		if tok == "--" {
			break
		}
		name, _, isFlag := splitFlagToken(tok)
		if isFlag && (name == "kind" || name == "supersedes") {
			return NewError(CodeUsage, "ant handover create sets --%s itself; use 'ant add' for anything else", name)
		}
	}

	store, _, err := a.requireInitialised()
	if err != nil {
		return err
	}
	previous, err := latestHandover(store)
	if err != nil {
		return err
	}

	addArgs := []string{"--kind", KindHandover}
	if previous != "" {
		addArgs = append(addArgs, "--supersedes", previous)
	}
	return a.Add(append(addArgs, args...))
}

// latestHandover returns the public id of the newest handover entry, or ""
// if there are none.
func latestHandover(store *Store) (string, error) {
	entries, err := store.ListEntries(ListFilter{Kind: KindHandover})
	if err != nil {
		return "", err
	}
	if len(entries) == 0 {
		return "", nil
	}
	return entries[0].PublicID, nil
}

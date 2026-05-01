package ant

import (
	"flag"
	"fmt"
	"io"
)

const bashCompletion = `# bash completion for ant
_ant_complete() {
    local cur prev
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    if [ "${COMP_CWORD}" -eq 1 ]; then
        COMPREPLY=( $(compgen -W "init config add show edit delete list recent search for export version completion" -- "${cur}") )
        return
    fi
    case "${prev}" in
        --kind)  COMPREPLY=( $(compgen -W "note adr pivot" -- "${cur}") ) ;;
        completion) COMPREPLY=( $(compgen -W "bash zsh" -- "${cur}") ) ;;
    esac
}
complete -F _ant_complete ant
`

const zshCompletion = `#compdef ant
_ant() {
    local -a commands kinds
    commands=(
        'init:initialise the .ant/ database'
        'config:show prefix, schema version, db path'
        'add:capture a new entry'
        'show:show full entry detail'
        'edit:update an existing entry'
        'delete:remove an entry'
        'list:list entries'
        'recent:show the most recent entries'
        'search:search entries by query'
        'for:show entries linked to an issue id'
        'export:render entries as markdown or JSON'
        'version:print the build version'
        'completion:print a shell completion script'
    )
    kinds=(note adr pivot)

    _arguments -C \
        '1: :->cmds' \
        '*::arg:->args'

    case ${state} in
        cmds) _describe 'ant command' commands ;;
        args)
            case ${words[1]} in
                completion) _values 'shell' bash zsh ;;
                add|edit) _arguments \
                    '--body[entry body or @file]:body:' \
                    '--title[entry title]:title:' \
                    '--kind[entry kind]:kind:(note adr pivot)' \
                    '--issue[linked issue id]:issue:' ;;
            esac
            ;;
    esac
}
_ant "$@"
`

// Completion handles 'ant completion {bash,zsh}'. Prints a completion
// script the user can source from their shell startup file.
func (a *App) Completion(args []string) error {
	fs := flag.NewFlagSet("completion", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: ant completion {bash|zsh}")
	}
	switch fs.Arg(0) {
	case "bash":
		_, err := io.WriteString(a.Stdout, bashCompletion)
		return err
	case "zsh":
		_, err := io.WriteString(a.Stdout, zshCompletion)
		return err
	default:
		return fmt.Errorf("unsupported shell %q (try bash or zsh)", fs.Arg(0))
	}
}

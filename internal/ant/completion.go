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

    local commands="init config add show edit append delete list recent search for foundation export version completion"
    local kinds="note adr pivot foundation"
    # Commands whose first positional argument is an entry id.
    local id_commands="show edit append delete export"

    if [ "${COMP_CWORD}" -eq 1 ]; then
        COMPREPLY=( $(compgen -W "${commands}" -- "${cur}") )
        return
    fi

    local cmd="${COMP_WORDS[1]}"

    # Flag value completions (apply regardless of subcommand).
    case "${prev}" in
        --kind)
            COMPREPLY=( $(compgen -W "${kinds}" -- "${cur}") )
            return
            ;;
    esac

    # 'completion' takes a shell name as its only argument.
    if [ "${cmd}" = "completion" ] && [ "${COMP_CWORD}" -eq 2 ]; then
        COMPREPLY=( $(compgen -W "bash zsh" -- "${cur}") )
        return
    fi

    # Entry id completion for commands that accept an id as a positional.
    # Only kicks in when the user is typing a non-flag token, so it doesn't
    # interfere with --flag value completions.
    if [[ "${cur}" != -* ]]; then
        for c in ${id_commands}; do
            if [ "${cmd}" = "${c}" ]; then
                local ids
                ids=$(ant list 2>/dev/null | grep -o '"id": *"[^"]*"' | sed 's/"id": *"//;s/"//')
                COMPREPLY=( $(compgen -W "${ids}" -- "${cur}") )
                return
            fi
        done
    fi
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
        'append:append to an existing entry, divided by a markdown rule'
        'delete:remove an entry'
        'list:list entries'
        'recent:show the most recent entries'
        'search:search entries by query'
        'for:show entries linked to an issue id'
        'foundation:show the project foundation entry'
        'export:render entries as markdown or JSON'
        'version:print the build version'
        'completion:print a shell completion script'
    )
    kinds=(note adr pivot foundation)

    # Pull entry ids out of 'ant list' JSON. Slim output is one "id" per
    # entry, so a quick grep/sed beats parsing JSON properly.
    _ant_entry_ids() {
        local -a ids
        ids=(${(f)"$(ant list 2>/dev/null | grep -o '"id": *"[^"]*"' | sed 's/"id": *"//;s/"//')"})
        compadd -a ids
    }

    _arguments -C \
        '1: :->cmds' \
        '*::arg:->args'

    case ${state} in
        cmds) _describe 'ant command' commands ;;
        args)
            case ${words[1]} in
                completion) _values 'shell' bash zsh ;;
                show|delete)
                    _ant_entry_ids
                    ;;
                edit)
                    if (( CURRENT == 2 )); then
                        _ant_entry_ids
                    else
                        _arguments \
                            '--body[new body, @file, or - for stdin]:body:' \
                            '--title[entry title]:title:' \
                            '--kind[entry kind]:kind:(note adr pivot foundation)' \
                            '--issue[linked issue id]:issue:' \
                            '--visual[open $EDITOR with the current body]' \
                            '--long[return the full record]'
                    fi
                    ;;
                append)
                    if (( CURRENT == 2 )); then
                        _ant_entry_ids
                    else
                        _arguments \
                            '--body[appended content, @file, or - for stdin]:body:' \
                            '--long[return the full record]'
                    fi
                    ;;
                export)
                    if [[ "${words[CURRENT]}" == -* ]]; then
                        _arguments \
                            '--kind[filter by kind]:kind:(note adr pivot foundation)' \
                            '--issue[filter by issue id]:issue:' \
                            '--since[created_at >= date]:since:' \
                            '--json[emit JSON instead of markdown]'
                    else
                        _ant_entry_ids
                    fi
                    ;;
                add)
                    _arguments \
                        '--body[entry body, @file, or - for stdin]:body:' \
                        '--title[entry title]:title:' \
                        '--kind[entry kind]:kind:(note adr pivot foundation)' \
                        '--issue[linked issue id]:issue:'
                    ;;
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
	fs.Usage = func() {
		fmt.Fprintln(a.Stderr, "usage: ant completion {bash|zsh}")
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return NewError(CodeValidationError, "usage: ant completion {bash|zsh}")
	}
	switch fs.Arg(0) {
	case "bash":
		_, err := io.WriteString(a.Stdout, bashCompletion)
		return err
	case "zsh":
		_, err := io.WriteString(a.Stdout, zshCompletion)
		return err
	default:
		return NewError(CodeValidationError, "unsupported shell %q (try bash or zsh)", fs.Arg(0))
	}
}

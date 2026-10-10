package main

const zshCompletion = `#compdef dotfiles

# Zsh completion script for dotfiles
# Install: cp _dotfiles ~/.oh-my-zsh/custom/completions/
#   — or — dotfiles -completion zsh > /path/to/completions/_dotfiles

_dotfiles() {
    local -a commands
    commands=(
        'status:show changes and unmanaged candidates'
        'diff:show changes against git repo'
        'add:move dotfiles into repo and create symlinks'
        'sync:pull remote changes, restore symlinks, commit and push'
        'update:stage, commit, and push changes'
        'doctor:check health of symlinks and permissions'
        'audit:scan managed dotfiles for sensitive credentials'
        'which:trace where an alias, function, or export is defined'
        'deps:audit external command dependencies across shell configs'
        'profile:manage machine configuration overlay profiles'
        'browse:interactive fuzzy finder with live preview and diffs'
        'fzf:interactive fuzzy finder with live preview and diffs'
        'snapshot:create or list tagged config checkpoints'
        'rollback:revert repo and symlinks to a saved snapshot'
        'edit:open matching dotfile directly in $EDITOR'
        'log:show commit history for dotfiles repo or specific file'
        'ignore:view or add patterns to .dotfilesignore'
        'guard:view or add sensitive guardrail patterns to .dotfilesguard'
        'clone:clone dotfiles repo and restore symlinks'
        'link:restore symlinks from repo contents'
        'clean:remove symlinks and restore original files'
    )

    _arguments -C \
        '(- *)'{-h,-help,--help}'[show help]' \
        '(- *)'{-version,--version}'[print version and exit]' \
        '(- *)'{-completion,--completion}'[print shell completion]:shell:(zsh bash)' \
        '(-a --all)'{-a,--all}'[show all files including ignored]' \
        '1: :->cmd' \
        '*:: :->args' && return

    case $state in
        cmd)
            _describe 'command' commands
            ;;
        args)
            case ${words[1]} in
                status)
                    _arguments \
                        '(-a --all)'{-a,--all}'[show all files including ignored]'
                    ;;
                diff)
                    _arguments \
                        '(--staged --cached)'{--staged,--cached}'[show staged changes]' \
                        '*:file:_files -g ".*"'
                    ;;
                add)
                    _arguments \
                        '(-f --force)'{-f,--force}'[override sensitive file guardrails]' \
                        '*:file:_files -g ".*"'
                    ;;
                doctor)
                    _arguments \
                        '(--fix -f)'{--fix,-f}'[automatically repair issues]'
                    ;;
                audit|edit|clean|link)
                    _files -g '.*'
                    ;;
                deps)
                    _arguments \
                        '(-s --strict)'{-s,--strict}'[fail if any dependencies are missing]'
                    ;;
                profile)
                    _arguments \
                        '(--clear)--clear[deactivate active profile]' \
                        '1:profile:'
                    ;;
                browse|fzf)
                    _arguments \
                        '(-p --print)'{-p,--print}'[print selected path instead of opening in editor]' \
                        '*:query:'
                    ;;
                snapshot)
                    _arguments \
                        '(-l --list)'{-l,--list}'[list available snapshots]' \
                        '(-d --delete)'{-d,--delete}'[delete snapshot]:snapshot:' \
                        '*:name:'
                    ;;
                rollback)
                    _arguments \
                        '1:snapshot:'
                    ;;
                which)
                    _arguments \
                        '1:name:'
                    ;;
                log)
                    _arguments \
                        '-n[number of commits to display]:count:' \
                        '*:file:_files -g ".*"'
                    ;;
                update|sync)
                    _arguments \
                        '-m[commit message]:message:' \
                        '(-f --force)'{-f,--force}'[override secret scanner]'
                    ;;
                clone)
                    _arguments \
                        '1:repository:_urls' \
                        '2:path:_directories'
                    ;;
            esac
            ;;
    esac
}

_dotfiles "$@"
`

const bashCompletion = `# Bash completion script for dotfiles
# Install: dotfiles -completion bash > /etc/bash_completion.d/dotfiles
#   — or — dotfiles -completion bash >> ~/.bashrc

_dotfiles() {
    local cur prev commands
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    commands="status diff add sync update doctor audit which deps profile browse fzf snapshot rollback edit log ignore guard clone link clean"

    if [[ ${COMP_CWORD} -eq 1 ]]; then
        if [[ "$cur" == -* ]]; then
            COMPREPLY=( $(compgen -W "-a --all -version --version -help --help -h -completion --completion" -- "$cur") )
        else
            COMPREPLY=( $(compgen -W "${commands}" -- "$cur") )
        fi
        return 0
    fi

    case "${COMP_WORDS[1]}" in
        -completion|--completion)
            COMPREPLY=( $(compgen -W "zsh bash" -- "$cur") )
            return 0
            ;;
        status)
            if [[ "$cur" == -* ]]; then
                COMPREPLY=( $(compgen -W "-a --all" -- "$cur") )
                return 0
            fi
            ;;
        diff)
            if [[ "$cur" == -* ]]; then
                COMPREPLY=( $(compgen -W "--staged --cached" -- "$cur") )
                return 0
            fi
            COMPREPLY=( $(compgen -f -X '!.*' -- "$cur") )
            return 0
            ;;
        add)
            if [[ "$cur" == -* ]]; then
                COMPREPLY=( $(compgen -W "-f --force" -- "$cur") )
                return 0
            fi
            COMPREPLY=( $(compgen -f -X '!.*' -- "$cur") )
            return 0
            ;;
        doctor)
            if [[ "$cur" == -* ]]; then
                COMPREPLY=( $(compgen -W "--fix -f" -- "$cur") )
                return 0
            fi
            ;;
        deps)
            if [[ "$cur" == -* ]]; then
                COMPREPLY=( $(compgen -W "-s --strict" -- "$cur") )
                return 0
            fi
            ;;
        profile)
            if [[ "$cur" == -* ]]; then
                COMPREPLY=( $(compgen -W "--clear" -- "$cur") )
                return 0
            fi
            ;;
        browse|fzf)
            if [[ "$cur" == -* ]]; then
                COMPREPLY=( $(compgen -W "-p --print" -- "$cur") )
                return 0
            fi
            ;;
        snapshot)
            if [[ "$cur" == -* ]]; then
                COMPREPLY=( $(compgen -W "-l --list -d --delete" -- "$cur") )
                return 0
            fi
            ;;
        audit|edit|clean|link|which|rollback)
            COMPREPLY=( $(compgen -f -X '!.*' -- "$cur") )
            return 0
            ;;
        log)
            if [[ "$cur" == -* ]]; then
                COMPREPLY=( $(compgen -W "-n" -- "$cur") )
                return 0
            fi
            COMPREPLY=( $(compgen -f -X '!.*' -- "$cur") )
            return 0
            ;;
        update|sync)
            if [[ "$cur" == -* ]]; then
                COMPREPLY=( $(compgen -W "-m -f --force" -- "$cur") )
                return 0
            fi
            ;;
        clone)
            COMPREPLY=( $(compgen -d -- "$cur") )
            return 0
            ;;
    esac
}

complete -F _dotfiles dotfiles
`

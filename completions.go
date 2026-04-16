package main

const zshCompletion = `#compdef dotfiles

# Zsh completion script for dotfiles
# Install: cp _dotfiles ~/.oh-my-zsh/custom/completions/
#   — or — dotfiles -completion zsh > /path/to/completions/_dotfiles

_dotfiles() {
    local -a commands
    commands=(
        'status:show changes and unmanaged candidates'
        'add:move dotfiles into repo and create symlinks'
        'update:stage, commit, and push changes'
        'clone:clone dotfiles repo and restore symlinks'
        'link:restore symlinks from repo contents'
        'clean:remove symlinks and restore original files'
    )

    _arguments -C \
        '(- *)'{-h,-help,--help}'[show help]' \
        '(- *)'{-version,--version}'[print version and exit]' \
        '(- *)'{-completion,--completion}'[print shell completion]:shell:(zsh bash)' \
        '1: :->cmd' \
        '*:: :->args' && return

    case $state in
        cmd)
            _describe 'command' commands
            ;;
        args)
            case ${words[1]} in
                add|clean|link)
                    _files -g '.*'
                    ;;
                update)
                    _arguments \
                        '-m[commit message]:message:'
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
    commands="status add update clone link clean"

    if [[ ${COMP_CWORD} -eq 1 ]]; then
        if [[ "$cur" == -* ]]; then
            COMPREPLY=( $(compgen -W "-version --version -help --help -h -completion --completion" -- "$cur") )
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
        add|clean|link)
            # Complete dotfiles (files starting with .)
            COMPREPLY=( $(compgen -f -X '!.*' -- "$cur") )
            return 0
            ;;
        update)
            if [[ "$cur" == -* ]]; then
                COMPREPLY=( $(compgen -W "-m" -- "$cur") )
                return 0
            fi
            ;;
        clone)
            # Complete directories for the path argument
            COMPREPLY=( $(compgen -d -- "$cur") )
            return 0
            ;;
    esac
}

complete -F _dotfiles dotfiles
`

package completion

func BashCompletion() string {
	return `#!/bin/bash

_cloud_swap() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    
    opts="add list update delete use assume config validate"
    creds=$(cloud-swap list 2>/dev/null | grep -oP '\\[\\d+\\] \\K[^ ]+' || true)
    
    case "${prev}" in
        cloud-swap)
            COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
            return 0
            ;;
        add)
            COMPREPLY=( $(compgen -W "--temp --name --access-key --secret-key --region" -- ${cur}) )
            return 0
            ;;
        use)
            COMPREPLY=( $(compgen -W "${creds} --export" -- ${cur}) )
            return 0
            ;;
        delete|update|validate)
            COMPREPLY=( $(compgen -W "${creds}" -- ${cur}) )
            return 0
            ;;
        assume)
            COMPREPLY=( $(compgen -W "${creds}" -- ${cur}) )
            return 0
            ;;
        config)
            COMPREPLY=( $(compgen -W "list set" -- ${cur}) )
            return 0
            ;;
        set)
            COMPREPLY=( $(compgen -W "default_region default_output fzf_enabled" -- ${cur}) )
            return 0
            ;;
    esac
    
    COMPREPLY=( $(compgen -W "${opts} ${creds}" -- ${cur}) )
    return 0
}

complete -F _cloud_swap cloud-swap
`
}

func InstallBash() string {
	return `# Add to ~/.bashrc or ~/.bash_profile:
source <(cloud-swap completion bash)

# Or install system-wide:
sudo cp completion/bash /etc/bash_completion.d/cloud-swap
`
}

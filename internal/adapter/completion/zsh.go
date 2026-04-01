package completion

func ZshCompletion() string {
	return `#!/usr/bin/env zsh

_cloud_swap() {
    local -a commands
    commands=(
        'add:add a new credential'
        'list:list all credentials'
        'update:update a credential'
        'delete:delete a credential'
        'use:activate a credential'
        'assume:assume an IAM role'
        'config:manage configuration'
        'validate:validate credentials'
    )
    
    local -a opts
    opts=(
        '--temp:add temporary credential'
        '--export:output environment variables'
        '--name:credential name (silent mode)'
        '--access-key:AWS access key ID'
        '--secret-key:AWS secret access key'
        '--region:AWS region'
    )
    
    local -a creds
    creds=(${(f)"$(cloud-swap list 2>/dev/null | grep -oP '\\[\\d+\\] \\K[^ ]+' || true)"})
    
    _describe 'command' commands
    _describe 'option' opts
    _describe 'credential' creds
    
    return 0
}

compdef _cloud_swap cloud-swap
`
}

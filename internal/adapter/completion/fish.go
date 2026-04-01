package completion

func FishCompletion() string {
	return `# Fish shell completion for cloud-swap

function __cloud-swap_get_credentials
    cloud-swap list 2>/dev/null | string replace -r '\[(\d+)\] (\S+).*' '$2'
end

complete -c cloud-swap -f -n '__fish_use_subcommand' -a 'add' -d 'add a new credential'
complete -c cloud-swap -f -n '__fish_use_subcommand' -a 'list' -d 'list all credentials'
complete -c cloud-swap -f -n '__fish_use_subcommand' -a 'update' -d 'update a credential'
complete -c cloud-swap -f -n '__fish_use_subcommand' -a 'delete' -d 'delete a credential'
complete -c cloud-swap -f -n '__fish_use_subcommand' -a 'use' -d 'activate a credential'
complete -c cloud-swap -f -n '__fish_use_subcommand' -a 'assume' -d 'assume an IAM role'
complete -c cloud-swap -f -n '__fish_use_subcommand' -a 'config' -d 'manage configuration'
complete -c cloud-swap -f -n '__fish_use_subcommand' -a 'validate' -d 'validate credentials'

complete -c cloud-swap -f -n '__fish_seen_subcommand_from add' -l temp -d 'add temporary credential'
complete -c cloud-swap -f -n '__fish_seen_subcommand_from add' -l name -d 'credential name (silent mode)'
complete -c cloud-swap -f -n '__fish_seen_subcommand_from add' -l access-key -d 'AWS access key ID'
complete -c cloud-swap -f -n '__fish_seen_subcommand_from add' -l secret-key -d 'AWS secret access key'
complete -c cloud-swap -f -n '__fish_seen_subcommand_from add' -l region -d 'AWS region'

complete -c cloud-swap -f -n '__fish_seen_subcommand_from use' -l export -d 'output environment variables'
complete -c cloud-swap -f -n '__fish_seen_subcommand_from use' -a '(__cloud-swap_get_credentials)'

complete -c cloud-swap -f -n '__fish_seen_subcommand_from delete' -a '(__cloud-swap_get_credentials)'
complete -c cloud-swap -f -n '__fish_seen_subcommand_from update' -a '(__cloud-swap_get_credentials)'
complete -c cloud-swap -f -n '__fish_seen_subcommand_from validate' -a '(__cloud-swap_get_credentials)'
complete -c cloud-swap -f -n '__fish_seen_subcommand_from assume' -a '(__cloud-swap_get_credentials)'

complete -c cloud-swap -f -n '__fish_seen_subcommand_from config' -a 'list set'
complete -c cloud-swap -f -n '__fish_seen_subcommand_from config set' -a 'default_region default_output fzf_enabled'
`
}

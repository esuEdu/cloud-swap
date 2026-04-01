package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/esuEdu/cloud-swap/internal/adapter/cli"
	"github.com/esuEdu/cloud-swap/internal/adapter/completion"
	"github.com/esuEdu/cloud-swap/internal/adapter/storage"
	"github.com/esuEdu/cloud-swap/internal/adapter/tui"
	"github.com/esuEdu/cloud-swap/internal/domain"
	"github.com/esuEdu/cloud-swap/internal/port"
	"github.com/esuEdu/cloud-swap/internal/usercase/credential"
)

func printUsage() {
	fmt.Println("cloud-swap <command>")
	fmt.Println("Commands:")
	fmt.Println("  add         - add a new credential")
	fmt.Println("  list        - list all credentials")
	fmt.Println("  update      - update a credential by name")
	fmt.Println("  delete      - delete a credential by name")
	fmt.Println("  use         - activate a credential by name")
	fmt.Println("  config      - manage configuration")
	fmt.Println("  validate    - validate credentials")
	fmt.Println("  backup      - backup credentials to file")
	fmt.Println("  restore     - restore credentials from file")
	fmt.Println("  sso-login   - AWS SSO login")
	fmt.Println("  completion  - generate shell completion")
	fmt.Println("  tui         - launch interactive TUI")
	fmt.Println("")
	fmt.Println("Flags:")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = printUsage

	tempFlag := flag.Bool("temp", false, "add a temporary credential (with expiration)")
	exportFlag := flag.Bool("export", false, "output environment variables for eval")

	nameFlag := flag.String("name", "", "credential name (silent mode)")
	accessKeyFlag := flag.String("access-key", "", "AWS access key ID (silent mode)")
	secretKeyFlag := flag.String("secret-key", "", "AWS secret access key (silent mode)")
	regionFlag := flag.String("region", "", "AWS region (silent mode)")

	passwordFlag := flag.String("password", "", "password for encryption/decryption")

	flag.Parse()

	if len(os.Args) < 2 {
		printUsage()
		return
	}

	var s port.Storage

	if *passwordFlag != "" {
		s = storage.NewEncryptedStorage(*passwordFlag)
	} else {
		s = storage.NewStorageRepo()
	}

	svc := credential.NewService(s)

	silent := *nameFlag != "" || *accessKeyFlag != "" || *secretKeyFlag != ""

	var c *cli.Cli
	if silent {
		c = cli.NewCliWithArgs(svc, silent, *nameFlag, *accessKeyFlag, *secretKeyFlag, *regionFlag)
	} else {
		c = cli.NewCli(svc)
	}

	switch os.Args[1] {
	case "add":
		if silent {
			os.Exit(0)
		}
		c.Add(*tempFlag)
	case "list":
		c.List()
	case "update":
		c.Update()
	case "delete":
		c.Delete()
	case "use":
		if len(os.Args) < 3 {
			fmt.Println("Usage: cloud-swap use <name>")
			return
		}
		c.Use(os.Args[2], *exportFlag)
	case "assume":
		if len(os.Args) < 3 {
			fmt.Println("Usage: cloud-swap assume <role-arn> [profile-name]")
			return
		}
		profile := "default"
		if len(os.Args) >= 4 {
			profile = os.Args[3]
		}
		c.Assume(os.Args[2], profile)
	case "config":
		handleConfig(os.Args[2:])
	case "sso-login":
		if len(os.Args) < 3 {
			fmt.Println("Usage: cloud-swap sso-login <profile-name>")
			return
		}
		handleSSOLogin(os.Args[2])
	case "completion":
		handleCompletion(os.Args[2:])
	case "tui":
		if err := tui.RunTUI(); err != nil {
			fmt.Printf("TUI error: %v\n", err)
		}
	case "backup":
		handleBackup(os.Args[2:])
	case "restore":
		handleRestore(os.Args[2:])
	case "validate":
		name := "default"
		if len(os.Args) >= 3 {
			name = os.Args[2]
		}
		if err := c.Validate(name); err != nil {
			fmt.Printf("Validation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Credentials are valid!")
	case "":
		c.Interactive()
	default:
		printUsage()
	}

}

func handleCompletion(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: cloud-swap completion <bash|zsh|fish>")
		return
	}

	switch args[0] {
	case "bash":
		fmt.Print(completion.BashCompletion())
	case "zsh":
		fmt.Print(completion.ZshCompletion())
	case "fish":
		fmt.Print(completion.FishCompletion())
	default:
		fmt.Printf("Unknown shell: %s (available: bash, zsh, fish)\n", args[0])
	}
}

func handleConfig(args []string) {
	configRepo := storage.NewConfigRepo()

	if len(args) == 0 || args[0] == "list" {
		cfg, err := configRepo.Load()
		if err != nil {
			fmt.Printf("Failed to load config: %v\n", err)
			return
		}
		fmt.Println("Current configuration:")
		fmt.Printf("  default_region: %s\n", cfg.DefaultRegion)
		fmt.Printf("  default_output: %s\n", cfg.DefaultOutput)
		fmt.Printf("  fzf_enabled:    %v\n", cfg.FzfEnabled)
		return
	}

	switch args[0] {
	case "set":
		if len(args) < 3 {
			fmt.Println("Usage: cloud-swap config set <key> <value>")
			fmt.Println("Keys: default_region, default_output, fzf_enabled")
			return
		}

		cfg, err := configRepo.Load()
		if err != nil {
			fmt.Printf("Failed to load config: %v\n", err)
			return
		}

		switch args[1] {
		case "default_region":
			cfg.DefaultRegion = args[2]
		case "default_output":
			cfg.DefaultOutput = args[2]
		case "fzf_enabled":
			cfg.FzfEnabled = args[2] == "true"
		default:
			fmt.Printf("Unknown key: %s\n", args[1])
			return
		}

		if err := configRepo.Save(cfg); err != nil {
			fmt.Printf("Failed to save config: %v\n", err)
			return
		}

		fmt.Println("Configuration updated!")
	default:
		fmt.Println("Usage: cloud-swap config [list|set <key> <value>]")
	}
}

func handleBackup(args []string) {
	storageRepo := storage.NewStorageRepo()

	if len(args) >= 1 && args[0] != "--all" {
		if err := storageRepo.Backup(args[0]); err != nil {
			fmt.Printf("Backup failed: %v\n", err)
			return
		}
		fmt.Printf("Backup saved to: %s\n", args[0])
		return
	}

	dir := "cloud-swap-backup"
	if len(args) >= 2 && args[0] == "--all" {
		dir = args[1]
	} else if len(args) >= 1 && args[0] == "--all" {
		dir = "cloud-swap-backup"
	}

	if err := storageRepo.BackupAll(dir); err != nil {
		fmt.Printf("Backup failed: %v\n", err)
		return
	}

	fmt.Printf("Full backup saved to: %s/\n", dir)
}

func handleRestore(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: cloud-swap restore <backup-file> [--all]")
		return
	}

	storageRepo := storage.NewStorageRepo()

	if len(args) >= 2 && args[1] == "--all" {
		if err := storageRepo.RestoreAll(args[0]); err != nil {
			fmt.Printf("Restore failed: %v\n", err)
			return
		}
		fmt.Printf("Full backup restored from: %s/\n", args[0])
		return
	}

	if err := storageRepo.Restore(args[0]); err != nil {
		fmt.Printf("Restore failed: %v\n", err)
		return
	}

	fmt.Printf("Credentials restored from: %s\n", args[0])
}

func handleSSOLogin(name string) {
	s := storage.NewStorageRepo()
	svc := credential.NewService(s)

	creds, err := svc.List()
	if err != nil {
		fmt.Printf("Failed to load credentials: %v\n", err)
		return
	}

	var selected *domain.Credential
	for _, c := range creds {
		if c.Name == name {
			selected = &c
			break
		}
	}

	if selected == nil {
		fmt.Printf("Credential not found: %s\n", name)
		return
	}

	if selected.Provider != "aws-sso" {
		fmt.Printf("Credential '%s' is not an AWS SSO profile (provider: %s)\n", name, selected.Provider)
		return
	}

	if err := credential.SSOLogin(selected); err != nil {
		fmt.Printf("SSO login failed: %v\n", err)
		return
	}
}

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/lola-os/lola-core/internal/config"
	"github.com/lola-os/lola-core/internal/vault"
)

var vaultCmd = &cobra.Command{
	Use:   "vault",
	Short: "Manage LOLA OS's encrypted key vault",
}

var vaultInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a new, empty vault",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if vault.Exists(cfg.Vault.Path) {
			return fmt.Errorf("a vault already exists at %s", cfg.Vault.Path)
		}
		pass, err := promptPassphrase("Set a vault passphrase: ", true)
		if err != nil {
			return err
		}
		v, err := vault.Create(cfg.Vault.Path, pass, vault.DefaultParams())
		if err != nil {
			return err
		}
		defer v.Close()
		fmt.Println(color.GreenString("✅ Vault created at " + cfg.Vault.Path))
		return nil
	},
}

var vaultAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add or update a private key in the vault, identified by <name>",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		vaultPass, err := promptPassphrase("Vault passphrase: ", false)
		if err != nil {
			return err
		}
		v, err := vault.OpenOrCreate(cfg.Vault.Path, vaultPass)
		if err != nil {
			return err
		}
		defer v.Close()

		secret, err := promptSecret("Private key (input hidden): ")
		if err != nil {
			return err
		}
		if err := v.Set(args[0], strings.TrimSpace(secret)); err != nil {
			return err
		}
		fmt.Println(color.GreenString(fmt.Sprintf("✅ Key %q stored.", args[0])))
		return nil
	},
}

var vaultListCmd = &cobra.Command{
	Use:   "list",
	Short: "List stored key names (never their values)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if !vault.Exists(cfg.Vault.Path) {
			fmt.Println("No vault found. Run `lola vault init` first.")
			return nil
		}
		vaultPass, err := promptPassphrase("Vault passphrase: ", false)
		if err != nil {
			return err
		}
		v, err := vault.Open(cfg.Vault.Path, vaultPass)
		if err != nil {
			return err
		}
		defer v.Close()
		names := v.List()
		if len(names) == 0 {
			fmt.Println("Vault is empty.")
			return nil
		}
		for _, n := range names {
			fmt.Println("  " + n)
		}
		return nil
	},
}

var vaultRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Delete a key from the vault",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		vaultPass, err := promptPassphrase("Vault passphrase: ", false)
		if err != nil {
			return err
		}
		v, err := vault.Open(cfg.Vault.Path, vaultPass)
		if err != nil {
			return err
		}
		defer v.Close()
		if err := v.Delete(args[0]); err != nil {
			return err
		}
		fmt.Println(color.GreenString(fmt.Sprintf("✅ Key %q removed.", args[0])))
		return nil
	},
}

func init() {
	vaultCmd.AddCommand(vaultInitCmd, vaultAddCmd, vaultListCmd, vaultRemoveCmd)
}

func promptPassphrase(prompt string, confirm bool) (string, error) {
	if env := os.Getenv("LOLA_VAULT_PASSPHRASE"); env != "" {
		return env, nil
	}
	pass, err := promptSecret(prompt)
	if err != nil {
		return "", err
	}
	if confirm {
		confirmPass, err := promptSecret("Confirm passphrase: ")
		if err != nil {
			return "", err
		}
		if pass != confirmPass {
			return "", fmt.Errorf("passphrases did not match")
		}
	}
	return pass, nil
}

// promptSecret reads a line from the terminal without echoing input, when
// stdin is a TTY; otherwise it falls back to a plain (echoed) read so the
// CLI remains usable in scripts/CI piping a passphrase via stdin.
func promptSecret(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	if term.IsTerminal(int(syscall.Stdin)) {
		bytePass, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return "", fmt.Errorf("reading passphrase: %w", err)
		}
		return string(bytePass), nil
	}
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("reading passphrase: %w", err)
	}
	return strings.TrimRight(line, "\r\n"), nil
}

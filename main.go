package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

// Version of the chicle tool
const Version = "0.0.1"

// UserConfig stores the configuration for each Git identity
type UserConfig struct {
	Name     string // Git user name
	Email    string // Git user email
	KeyPath  string // Path to the SSH key file
	IsGlobal bool   // Whether this is a global identity
}

// Configs stores both global and local configurations
type Configs struct {
	Global map[string]UserConfig
	Local  map[string]UserConfig
}

var configs Configs

// loadConfigs reads the configuration file and populates configs
func loadConfigs() error {
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".chicle_config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// If the file doesn't exist, initialize empty maps
			configs = Configs{
				Global: make(map[string]UserConfig),
				Local:  make(map[string]UserConfig),
			}
			return nil
		}
		return err
	}

	// Unmarshal JSON data into configs
	return json.Unmarshal(data, &configs)
}

// saveConfigs writes the current configs to the configuration file
func saveConfigs() error {
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".chicle_config.json")

	// Marshal configs into JSON
	data, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return err
	}

	// Write JSON data to the configuration file
	return os.WriteFile(configPath, data, 0644)
}

func main() {
	// Load existing configurations
	err := loadConfigs()
	if err != nil {
		log.Fatal(err)
	}

	// Define the CLI application
	app := &cli.App{
		Name:    "chicle",
		Usage:   "Git User Manager - Platform-agnostic tool for managing multiple Git identities",
		Version: Version,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "Print the version",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "create",
				Usage: "Create a new SSH key and Git identity or add an existing key",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "alias",
						Aliases:  []string{"a"},
						Usage:    "Alias for the Git identity",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "email",
						Aliases:  []string{"e"},
						Usage:    "Email address for the SSH key",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "name",
						Aliases:  []string{"n"},
						Usage:    "Git user name",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "key",
						Aliases: []string{"k"},
						Usage:   "Path to an existing SSH key (optional)",
					},
					&cli.BoolFlag{
						Name:    "global",
						Aliases: []string{"g"},
						Usage:   "Set the identity globally (optional)",
					},
				},
				Action: createIdentity,
			},
			{
				Name:      "switch",
				Usage:     "Switch Git user",
				ArgsUsage: "ALIAS",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "global",
						Aliases: []string{"g"},
						Usage:   "Switch identity globally (optional)",
					},
				},
				Action: switchUser,
			},
			{
				Name:   "list",
				Usage:  "List all stored Git identities",
				Action: listIdentities,
			},
			{
				Name:      "delete",
				Usage:     "Delete a Git identity",
				ArgsUsage: "ALIAS",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "global",
						Aliases: []string{"g"},
						Usage:   "Delete a global identity (optional)",
					},
				},
				Action: deleteIdentity,
			},
		},
	}

	// Run the CLI application
	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// isGitRepository checks if the current directory is a Git repository
func isGitRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	err := cmd.Run()
	return err == nil
}

// createIdentity handles the creation of a new Git identity
func createIdentity(c *cli.Context) error {
	alias := c.String("alias")
	email := c.String("email")
	name := c.String("name")
	existingKey := c.String("key")
	isGlobal := c.Bool("global")

	if !isGlobal && !isGitRepository() {
		return fmt.Errorf("not in a Git repository. Use -g flag to create a global identity")
	}

	// Check if the alias already exists in either global or local configs
	if _, exists := configs.Global[alias]; exists {
		return fmt.Errorf("alias '%s' already exists as a global identity", alias)
	}
	if _, exists := configs.Local[alias]; exists {
		return fmt.Errorf("alias '%s' already exists as a local identity", alias)
	}

	var keyPath string

	if existingKey != "" {
		// Check if the provided key exists and is valid
		if err := validateExistingKey(existingKey); err != nil {
			return err
		}
		keyPath = existingKey
	} else {
		// Generate a new SSH key
		homeDir, _ := os.UserHomeDir()
		keyPath = filepath.Join(homeDir, ".ssh", fmt.Sprintf("id_rsa_%s", alias))
		if err := generateSSHKey(email, keyPath); err != nil {
			return err
		}
	}

	// Store the new configuration
	newConfig := UserConfig{
		Name:     name,
		Email:    email,
		KeyPath:  keyPath,
		IsGlobal: isGlobal,
	}

	if isGlobal {
		configs.Global[alias] = newConfig
	} else {
		configs.Local[alias] = newConfig
	}

	// Save the updated configurations
	if err := saveConfigs(); err != nil {
		return fmt.Errorf("error saving configuration: %v", err)
	}

	// Set Git configs
	gitConfigCmd := "git"
	if isGlobal {
		gitConfigCmd += " config --global"
	} else {
		gitConfigCmd += " config"
	}

	configs := [][]string{
		{"user.name", name},
		{"user.email", email},
		{"core.sshCommand", fmt.Sprintf("ssh -i %s", keyPath)},
	}

	for _, cfg := range configs {
		cmd := exec.Command("sh", "-c", fmt.Sprintf("%s %s \"%s\"", gitConfigCmd, cfg[0], cfg[1]))
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error setting Git config %s: %v\n%s", cfg[0], err, output)
		}
	}

	fmt.Printf("Identity created for alias '%s' with email '%s' and name '%s'\n", alias, email, name)
	fmt.Printf("SSH key path: %s\n", keyPath)
	if existingKey == "" {
		fmt.Println("Remember to add this key to your Git hosting service.")
	}
	if isGlobal {
		fmt.Println("This identity has been set globally.")
	} else {
		fmt.Println("This identity has been set for the current Git repository.")
	}
	return nil
}

// switchUser handles switching to a different Git identity
func switchUser(c *cli.Context) error {
	if c.NArg() < 1 {
		return fmt.Errorf("missing alias argument. Usage: chicle switch ALIAS")
	}
	alias := c.Args().First()
	isGlobal := c.Bool("global")

	if !isGlobal && !isGitRepository() {
		return fmt.Errorf("not in a Git repository. Use -g flag to switch globally")
	}

	var config UserConfig
	var ok bool

	if isGlobal {
		config, ok = configs.Global[alias]
	} else {
		config, ok = configs.Local[alias]
	}

	if !ok {
		scopeType := "local"
		if isGlobal {
			scopeType = "global"
		}
		return fmt.Errorf("no %s identity found for alias '%s'", scopeType, alias)
	}

	// Clear existing SSH keys from the agent
	cmd := exec.Command("ssh-add", "-D")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error clearing SSH keys: %v\n%s", err, output)
	}

	// Add the new SSH key
	cmd = exec.Command("ssh-add", config.KeyPath)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error adding SSH key: %v\n%s", err, output)
	}

	// Set Git configs
	gitConfigCmd := "git"
	if isGlobal {
		gitConfigCmd += " config --global"
	} else {
		gitConfigCmd += " config"
	}

	configs := [][]string{
		{"user.name", config.Name},
		{"user.email", config.Email},
		{"core.sshCommand", fmt.Sprintf("ssh -i %s", config.KeyPath)},
	}

	for _, cfg := range configs {
		cmd := exec.Command("sh", "-c", fmt.Sprintf("%s %s \"%s\"", gitConfigCmd, cfg[0], cfg[1]))
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error setting Git config %s: %v\n%s", cfg[0], err, output)
		}
	}

	fmt.Printf("Switched to user: %s (%s)\n", config.Name, config.Email)
	if isGlobal {
		fmt.Println("This identity has been set globally.")
	} else {
		fmt.Println("This identity has been set for the current Git repository.")
	}
	fmt.Println("This configuration will work with any Git hosting service.")
	return nil
}

// deleteIdentity handles the deletion of a Git identity
func deleteIdentity(c *cli.Context) error {
	if c.NArg() < 1 {
		return fmt.Errorf("missing alias argument. Usage: chicle delete ALIAS")
	}
	alias := c.Args().First()
	isGlobal := c.Bool("global")

	var config UserConfig
	var ok bool

	if isGlobal {
		config, ok = configs.Global[alias]
		if !ok {
			return fmt.Errorf("no global identity found for alias '%s'", alias)
		}
		delete(configs.Global, alias)
	} else {
		config, ok = configs.Local[alias]
		if !ok {
			return fmt.Errorf("no local identity found for alias '%s'", alias)
		}
		delete(configs.Local, alias)
	}

	// Save the updated configurations
	if err := saveConfigs(); err != nil {
		return fmt.Errorf("error saving configuration: %v", err)
	}

	fmt.Printf("Identity '%s' (%s) has been deleted.\n", alias, config.Email)
	fmt.Println("Note: The associated SSH key file was not deleted. You may want to remove it manually if it's no longer needed.")

	return nil
}

// listIdentities displays all stored Git identities
func listIdentities(c *cli.Context) error {
	fmt.Println("Global Git Identities:")
	for alias, config := range configs.Global {
		fmt.Printf("- Alias: %s\n  Name: %s\n  Email: %s\n  Key: %s\n\n", alias, config.Name, config.Email, config.KeyPath)
	}

	fmt.Println("Local Git Identities:")
	for alias, config := range configs.Local {
		fmt.Printf("- Alias: %s\n  Name: %s\n  Email: %s\n  Key: %s\n\n", alias, config.Name, config.Email, config.KeyPath)
	}
	return nil
}

// validateExistingKey checks if the provided key exists and is valid
func validateExistingKey(keyPath string) error {
	// Check if the file exists
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return fmt.Errorf("the specified key file does not exist: %s", keyPath)
	}

	// Check if the key is valid
	cmd := exec.Command("ssh-keygen", "-l", "-f", keyPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("the specified key is not a valid SSH key: %v", err)
	}

	return nil
}

// generateSSHKey creates a new SSH key
func generateSSHKey(email, keyPath string) error {
	cmd := exec.Command("ssh-keygen", "-t", "rsa", "-b", "4096", "-C", email, "-f", keyPath, "-N", "")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error creating SSH key: %v\n%s", err, output)
	}
	return nil
}

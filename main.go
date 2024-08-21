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
const Version = "0.0.3+1"

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
var verbose bool

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
				Name:        "verbose",
				Aliases:     []string{"V"},
				Usage:       "Enable verbose output",
				Destination: &verbose,
			},
		},
		Commands: []*cli.Command{
			createCommand(),
			switchCommand(),
			deleteCommand(),
			listCommand(),
			configCommand(),
		},
	}

	// Run the CLI application
	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func createCommand() *cli.Command {
	return &cli.Command{
		Name:    "create",
		Aliases: []string{"add"},
		Usage:   "Create a new SSH key and Git identity or add an existing key",
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
	}
}

func switchCommand() *cli.Command {
	return &cli.Command{
		Name:    "switch",
		Aliases: []string{"use"},
		Usage:   "Switch Git user",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "global",
				Aliases: []string{"g"},
				Usage:   "Switch identity globally",
			},
		},
		Action: func(c *cli.Context) error {
			isGlobal := false
			for _, arg := range os.Args {
				if arg == "--global" || arg == "-g" {
					isGlobal = true
					break
				}
			}
			//fmt.Printf("Debug: Inside switch command action, isGlobal: %v\n", isGlobal)
			//fmt.Printf("Debug: All args: %v\n", os.Args)
			return switchUser(c, isGlobal)
		},
	}
}

func deleteCommand() *cli.Command {
	return &cli.Command{
		Name:    "delete",
		Aliases: []string{"remove"},
		Usage:   "Delete a Git identity",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "global",
				Aliases: []string{"g"},
				Usage:   "Delete a global identity",
			},
		},
		Action: deleteIdentity,
	}
}

func listCommand() *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "List all stored Git identities",
		Action:  listIdentities,
	}
}

func configCommand() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "Configure default behaviors",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "always-global",
				Usage: "Set operations to always be global by default",
			},
		},
		Action: configureDefaults,
	}
}

func createIdentity(c *cli.Context) error {
	alias := c.String("alias")
	email := c.String("email")
	name := c.String("name")
	existingKey := c.String("key")
	isGlobal := c.Bool("global")

	if verbose {
		log.Printf("Creating identity - Alias: %s, Email: %s, Name: %s, Global: %v\n", alias, email, name, isGlobal)
	}

	if !isGlobal && !isGitRepository() {
		return cli.NewExitError("Not in a Git repository. Use --global flag to create a global identity or navigate to a Git repository.", 1)
	}

	// Check if the alias already exists in either global or local configs
	if _, exists := configs.Global[alias]; exists {
		return cli.NewExitError(fmt.Sprintf("Alias '%s' already exists as a global identity. Please choose a different alias.", alias), 1)
	}
	if _, exists := configs.Local[alias]; exists {
		return cli.NewExitError(fmt.Sprintf("Alias '%s' already exists as a local identity. Please choose a different alias.", alias), 1)
	}

	var keyPath string

	if existingKey != "" {
		// Check if the provided key exists and is valid
		if err := validateExistingKey(existingKey); err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
		keyPath = existingKey
	} else {
		// Generate a new SSH key
		homeDir, _ := os.UserHomeDir()
		keyPath = filepath.Join(homeDir, ".ssh", fmt.Sprintf("id_rsa_%s", alias))
		if err := generateSSHKey(email, keyPath); err != nil {
			return cli.NewExitError(err.Error(), 1)
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
		return cli.NewExitError(fmt.Sprintf("Error saving configuration: %v", err), 1)
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
			return cli.NewExitError(fmt.Sprintf("Error setting Git config %s: %v\n%s", cfg[0], err, output), 1)
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

func switchUser(c *cli.Context, isGlobal bool) error {
	args := c.Args()
	if args.Len() < 1 {
		return cli.NewExitError("Missing alias argument. Usage: chicle switch [--global] ALIAS", 1)
	}
	alias := args.First()

	//fmt.Printf("Debug: switchUser called with alias: %s, isGlobal: %v\n", alias, isGlobal)
	//fmt.Printf("Debug: All args: %v\n", os.Args)
	//fmt.Printf("Debug: Global configs: %+v\n", configs.Global)
	//fmt.Printf("Debug: Local configs: %+v\n", configs.Local)

	var config UserConfig
	var ok bool

	if isGlobal {
		//fmt.Println("Debug: Checking global configs")
		config, ok = configs.Global[alias]
	} else {
		//fmt.Println("Debug: Checking local configs")
		config, ok = configs.Local[alias]
	}

	//fmt.Printf("Debug: Config found: %v, Config: %+v\n", ok, config)

	if !ok {
		scopeType := map[bool]string{true: "global", false: "local"}[isGlobal]
		return cli.NewExitError(fmt.Sprintf("No %s identity found for alias '%s'. Use 'chicle list' to see available identities.", scopeType, alias), 1)
	}

	// Clear existing SSH keys from the agent
	cmd := exec.Command("ssh-add", "-D")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Error clearing SSH keys: %v\n%s", err, output), 1)
	}

	// Add the new SSH key
	cmd = exec.Command("ssh-add", config.KeyPath)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Error adding SSH key: %v\n%s", err, output), 1)
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
			return cli.NewExitError(fmt.Sprintf("Error setting Git config %s: %v\n%s", cfg[0], err, output), 1)
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

func deleteIdentity(c *cli.Context) error {
	if c.NArg() < 1 {
		return cli.NewExitError("Missing alias argument. Usage: chicle delete [--global] ALIAS", 1)
	}
	alias := c.Args().First()
	isGlobal := c.Bool("global")

	if verbose {
		log.Printf("Deleting identity - Alias: %s, Global: %v\n", alias, isGlobal)
	}

	var config UserConfig
	var ok bool

	if isGlobal {
		config, ok = configs.Global[alias]
		if !ok {
			return cli.NewExitError(fmt.Sprintf("No global identity found for alias '%s'. Use 'chicle list' to see available identities.", alias), 1)
		}
		delete(configs.Global, alias)
	} else {
		config, ok = configs.Local[alias]
		if !ok {
			return cli.NewExitError(fmt.Sprintf("No local identity found for alias '%s'. Use 'chicle list' to see available identities.", alias), 1)
		}
		delete(configs.Local, alias)
	}

	// Save the updated configurations
	if err := saveConfigs(); err != nil {
		return cli.NewExitError(fmt.Sprintf("Error saving configuration: %v", err), 1)
	}

	fmt.Printf("Identity '%s' (%s) has been deleted.\n", alias, config.Email)
	fmt.Println("Note: The associated SSH key file was not deleted. You may want to remove it manually if it's no longer needed.")

	return nil
}

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

func configureDefaults(c *cli.Context) error {
	alwaysGlobal := c.Bool("always-global")

	if verbose {
		log.Printf("Configuring defaults - Always Global: %v\n", alwaysGlobal)
	}

	// Implementation of configuration storage and application
	// This is a placeholder and should be implemented based on how you want to store and use these configurations
	fmt.Printf("Default behavior set: Always Global = %v\n", alwaysGlobal)
	fmt.Println("Note: This feature is not fully implemented yet.")

	return nil
}

// isGitRepository checks if the current directory is a Git repository
func isGitRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	err := cmd.Run()
	return err == nil
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

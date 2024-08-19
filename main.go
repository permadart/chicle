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

// UserConfig stores the configuration for each Git identity
type UserConfig struct {
	Name    string // Git user name
	Email   string // Git user email
	KeyPath string // Path to the SSH key file
}

// userConfigs is a map to store multiple user configurations
var userConfigs map[string]UserConfig

// loadConfigs reads the configuration file and populates userConfigs
func loadConfigs() error {
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".gum_config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// If the file doesn't exist, initialize an empty map
			userConfigs = make(map[string]UserConfig)
			return nil
		}
		return err
	}

	// Unmarshal JSON data into userConfigs
	return json.Unmarshal(data, &userConfigs)
}

// saveConfigs writes the current userConfigs to the configuration file
func saveConfigs() error {
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".gum_config.json")

	// Marshal userConfigs into JSON
	data, err := json.MarshalIndent(userConfigs, "", "  ")
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
		Name:  "gum",
		Usage: "Git User Manager - Platform-agnostic tool for managing multiple Git identities",
		Commands: []*cli.Command{
			{
				Name:  "create",
				Usage: "Create a new SSH key and Git identity",
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
				},
				Action: createIdentity,
			},
			{
				Name:      "switch",
				Usage:     "Switch Git user",
				ArgsUsage: "ALIAS",
				Action:    switchUser,
			},
			{
				Name:   "list",
				Usage:  "List all stored Git identities",
				Action: listIdentities,
			},
		},
	}

	// Run the CLI application
	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// createIdentity handles the creation of a new Git identity
func createIdentity(c *cli.Context) error {
	alias := c.String("alias")
	email := c.String("email")
	name := c.String("name")

	homeDir, _ := os.UserHomeDir()
	keyPath := filepath.Join(homeDir, ".ssh", fmt.Sprintf("id_rsa_%s", alias))

	// Generate a new SSH key
	cmd := exec.Command("ssh-keygen", "-t", "rsa", "-b", "4096", "-C", email, "-f", keyPath, "-N", "")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error creating SSH key: %v\n%s", err, output)
	}

	// Store the new configuration
	userConfigs[alias] = UserConfig{
		Name:    name,
		Email:   email,
		KeyPath: keyPath,
	}

	// Save the updated configurations
	err = saveConfigs()
	if err != nil {
		return fmt.Errorf("error saving configuration: %v", err)
	}

	fmt.Printf("Identity created for alias '%s' with email '%s' and name '%s'\n", alias, email, name)
	fmt.Printf("SSH key created successfully: %s\n", keyPath)
	fmt.Println("Remember to add this key to your Git hosting service.")
	return nil
}

// switchUser handles switching to a different Git identity
func switchUser(c *cli.Context) error {
	if c.NArg() < 1 {
		return fmt.Errorf("missing alias argument. Usage: gum switch ALIAS")
	}
	alias := c.Args().First()

	config, ok := userConfigs[alias]
	if !ok {
		return fmt.Errorf("no identity found for alias '%s'", alias)
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
	configs := [][]string{
		{"user.name", config.Name},
		{"user.email", config.Email},
		{"core.sshCommand", fmt.Sprintf("ssh -i %s", config.KeyPath)},
	}

	for _, cfg := range configs {
		cmd = exec.Command("git", "config", "--global", cfg[0], cfg[1])
		output, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error setting Git config %s: %v\n%s", cfg[0], err, output)
		}
	}

	fmt.Printf("Switched to user: %s (%s)\n", config.Name, config.Email)
	fmt.Println("This configuration will work with any Git hosting service.")
	return nil
}

// listIdentities displays all stored Git identities
func listIdentities(c *cli.Context) error {
	fmt.Println("Stored Git Identities:")
	for alias, config := range userConfigs {
		fmt.Printf("- Alias: %s\n  Name: %s\n  Email: %s\n  Key: %s\n\n", alias, config.Name, config.Email, config.KeyPath)
	}
	return nil
}

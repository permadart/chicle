package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "gum",
		Usage: "Git User Manager - Platform-agnostic tool for managing multiple Git identities",
		Description: `gum is a platform-agnostic tool that works with any Git hosting service, 
including but not limited to GitHub, GitLab, and Bitbucket. It manages your 
local Git configurations and SSH keys, allowing you to switch between different 
Git identities easily.`,
		Commands: []*cli.Command{
			{
				Name:  "create",
				Usage: "Create a new SSH key",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "email",
						Aliases:  []string{"e"},
						Usage:    "Email address for the SSH key",
						Required: true,
					},
				},
				Action: createSSHKey,
			},
			{
				Name:  "switch",
				Usage: "Switch Git user",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Aliases:  []string{"n"},
						Usage:    "Git user name",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "email",
						Aliases:  []string{"e"},
						Usage:    "Git user email",
						Required: true,
					},
				},
				Action: switchUser,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func createSSHKey(c *cli.Context) error {
	email := c.String("email")
	homeDir, _ := os.UserHomeDir()
	keyPath := filepath.Join(homeDir, ".ssh", fmt.Sprintf("id_rsa_%s", email))

	cmd := exec.Command("ssh-keygen", "-t", "rsa", "-b", "4096", "-C", email, "-f", keyPath, "-N", "")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error creating SSH key: %v\n%s", err, output)
	}

	fmt.Printf("SSH key created successfully: %s\n", keyPath)
	fmt.Println("Remember to add this key to your Git hosting service.")
	return nil
}

func switchUser(c *cli.Context) error {
	name := c.String("name")
	email := c.String("email")
	homeDir, _ := os.UserHomeDir()
	keyPath := filepath.Join(homeDir, ".ssh", fmt.Sprintf("id_rsa_%s", email))

	// Add SSH key
	cmd := exec.Command("ssh-add", keyPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error adding SSH key: %v\n%s", err, output)
	}

	// Check for existing platform-specific configurations
	existingConfig, err := exec.Command("git", "config", "--global", "--get-regexp", "^url\\..*\\.insteadOf").Output()
	if err == nil && len(existingConfig) > 0 {
		fmt.Println("Warning: Detected existing platform-specific Git configurations. These will not be modified.")
	}

	// Set Git configs
	configs := [][]string{
		{"user.name", name},
		{"user.email", email},
		{"core.sshCommand", fmt.Sprintf("ssh -i %s", keyPath)},
	}

	for _, config := range configs {
		cmd = exec.Command("git", "config", "--global", config[0], config[1])
		output, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error setting Git config %s: %v\n%s", config[0], err, output)
		}
	}

	fmt.Printf("Switched to user: %s (%s)\n", name, email)
	fmt.Println("This configuration will work with any Git hosting service.")
	return nil
}

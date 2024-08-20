# chicle (Git User Manager)

chicle is a platform-agnostic command-line tool for managing multiple Git identities. It simplifies the process of switching between different Git user configurations and SSH keys, making it easier to work with multiple Git accounts across various hosting services like GitHub, GitLab, and Bitbucket.

## Features

- Create SSH keys for different Git identities
- Add existing SSH keys to new identities
- Switch between Git user configurations easily
- Works with any Git hosting service
- Manages both global and local Git configurations and SSH keys
- Ensures unique aliases across global and local identities
- Supports both global and repository-specific identity management
- Version checking functionality

## Installation

### Using Homebrew

You can install chicle using Homebrew:

```bash
brew install permadart/chicle/chicle
```

### From Source

If you prefer to install from source or Homebrew is not available:

1. Ensure you have Go installed on your system.
2. Clone the repository:
   ```bash
   git clone https://github.com/permadart/chicle.git
   ```
3. Navigate to the project directory:
   ```bash
   cd chicle
   ```
4. Build the project:
   ```bash
   go build -o chicle
   ```
5. (Optional) Move the binary to a location in your PATH:
   ```bash
   sudo mv chicle /usr/local/bin/
   ```

## Usage

### Create a new SSH key and identity (globally)

```bash
chicle create --alias user1 --email user@example.com --name "John Doe" --global
```

### Create a new SSH key and identity (for the current repository)

```bash
chicle create --alias user1 --email user@example.com --name "John Doe"
```

### Add an existing SSH key to a new identity (globally)

```bash
chicle create --alias user2 --email user2@example.com --name "Jane Doe" --key ~/.ssh/id_rsa_user2 --global
```

### Add an existing SSH key to a new identity (for the current repository)

```bash
chicle create --alias user2 --email user2@example.com --name "Jane Doe" --key ~/.ssh/id_rsa_user2
```

### Switch Git user (globally)

```bash
chicle switch user1 --global
```

### Switch Git user (for the current repository)

```bash
chicle switch user1
```

### Delete a Git identity (globally)

```bash
chicle delete user1 --global
```

### Delete a Git identity (local)

```bash
chicle delete user1
```

### List all identities

```bash
chicle list
```

### Check chicle version

```bash
chicle --version
```
or
```bash
chicle -v
```

## Important Notes

- When creating or switching to a local identity (without the `--global` flag), you must be inside a Git repository. If you're not in a Git repository, chicle will return an error and prompt you to use the `--global` flag.
- Aliases must be unique across both global and local identities. chicle will prevent you from creating an identity with an alias that already exists in either scope.
- The `list` command shows both global and local identities separately for better clarity.
- When switching identities, chicle clears existing SSH keys from the agent before adding the new one to ensure a clean switch.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
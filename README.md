# gum (Git User Manager)

gum is a platform-agnostic command-line tool for managing multiple Git identities. It simplifies the process of switching between different Git user configurations and SSH keys, making it easier to work with multiple Git accounts across various hosting services like GitHub, GitLab, and Bitbucket.

## Features

- Create SSH keys for different Git identities
- Switch between Git user configurations easily
- Works with any Git hosting service
- Manages global Git configurations and SSH keys

## Installation

### Using Homebrew

You can install gum using Homebrew:

```bash
brew install permadart/gum/gum
```

### From Source

If you prefer to install from source or Homebrew is not available:

1. Ensure you have Go installed on your system.
2. Clone the repository:
   ```bash
   git clone https://github.com/permadart/gum.git
   ```
3. Navigate to the project directory:
   ```bash
   cd gum
   ```
4. Build the project:
   ```bash
   go build -o gum
   ```
5. (Optional) Move the binary to a location in your PATH:
   ```bash
   sudo mv gum /usr/local/bin/
   ```

## Usage

### Create a new SSH key

```bash
gum create --email user@example.com
```

### Switch Git user

```bash
gum switch --name "John Doe" --email user@example.com
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
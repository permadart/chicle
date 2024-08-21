# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.0.3] - 2024-08-20

### Added
- Verbose mode with --verbose flag for detailed command execution information
- Command aliases for improved usability:
  - `add` for `create`
  - `use` for `switch`
  - `remove` for `delete`
  - `ls` for `list`
- New configuration command to set default behaviors (e.g., --always-global)
- CONTRIBUTING.md file with development guidelines

### Changed
- Improved flag parsing and error handling for all commands
- Enhanced error messages with contextual information and usage hints
- Updated README with clearer explanations of local vs global operations and more examples
- Restructured commands into separate functions for better code organization

### Fixed
- Resolved issue with global flag recognition in the switch command

### Improved
- Input validation across all commands
- User guidance in error messages, suggesting correct usage and next steps

## [0.0.2] - 2024-08-20

### Changed
- Improved error handling and debug logging
- Updated version number

## [0.0.1] - 2024-08-20

### Added
- Initial release of chicle
- Create command for generating SSH keys and creating Git identities
- Switch command for changing Git user configurations
- Delete command to remove Git identities by alias
- List command to display all stored identities
- Persistent storage of user configurations
- Global and local identity management
- Version checking functionality (--version flag)
- README.md with usage instructions
- MIT License

[0.0.3]: https://github.com/permadart/chicle/compare/v0.0.2...v0.0.3
[0.0.2]: https://github.com/permadart/chicle/compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/permadart/chicle/releases/tag/v0.0.1
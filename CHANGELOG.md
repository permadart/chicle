# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.0.1] - 2024-08-20

### Added
- Initial release of chicle
- Create command for generating SSH keys and creating Git identities
  - Support for adding existing SSH keys when creating a new identity
  - New --key flag in the create command to specify an existing SSH key
  - Validation of existing SSH keys before adding them to an identity
- Switch command for changing Git user configurations
- Delete command to remove Git identities by alias
- List command to display all stored identities
- Persistent storage of user configurations
- Alias support for creating and switching between Git identities
- Global and local identity management
- Unique alias protection across global and local scopes
- Version checking functionality (--version flag)
- README.md with comprehensive usage instructions
- CHANGELOG.md to track changes
- MIT License

### Changed
- The tool is now called chicle (previously had a different name)

[0.0.1]: https://github.com/permadart/chicle/releases/tag/v0.0.1
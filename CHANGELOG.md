# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.0.1+2] - 2024-08-19

### Added

- Alias support for creating and switching between Git identities
- New list command to display all stored identities
- Persistent storage of user configurations

### Changed

- create command now requires an alias and stores the configuration
- switch command now works with aliases instead of individual parameters

### Fixed

## [0.0.1] - 2024-08-19

### Added
- Initial release of gum
- Create command for generating SSH keys
- Switch command for changing Git user configurations
- README.md with usage instructions
- CHANGELOG.md to track changes
- MIT License

[Unreleased]: https://github.com/permadart/gum/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/permadart/gum/releases/tag/v1.0.0
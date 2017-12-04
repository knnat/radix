# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- Alphabetical sorting.

### Changed
- Sorting is not made automatically anymore.
- Characters in the printed tree.
- README file.

### Fixed
- CHANGELOG code citations are shown as code instead of ordinary text.
- README typo.
- Tests are now run in the correct order.
- `Tree.Add` now correctly adds a node as a prefix.

## [0.2.1] - 2017-11-11
### Added
- Benchmark flag.

### Fixed
- `goimports` installation missing.

## [0.2.0] - 2017-11-08
### Added
- Specific methods for inserting and deleting a node and also return the tree.

### Changed
- Enhance basic operations' algorithms.
- Remove basic operations' return value.

## [0.1.3] - 2017-11-01
### Added
- Some use cases to the tests table.

### Changed
- Variable name `Node.Val` to `Node.Value`.

### Fixed
- GetByRune method returning a non-nil map even when not matching the string.

## [0.1.2] - 2017-11-01
### Changed
- Update this file to use "changelog" in lieu of "change log".

### Fixed
- Out of range bug when dynamically looking up a string.

## [0.1.1] - 2017-10-31
### Changed
- README file.

### Fixed
- Output examples.

## 0.1.0 - 2017-10-31
### Added
- This changelog file.
- README file.
- MIT License.
- Travis CI configuration file and scripts.
- Git ignore file.
- Editorconfig file.
- This package's source code, including examples and tests.
- Go dep files.

[Unreleased]: https://github.com/gbrlsnchs/patricia/compare/v0.2.1...HEAD
[0.2.1]: https://github.com/gbrlsnchs/patricia/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/gbrlsnchs/patricia/compare/v0.1.3...v0.2.0
[0.1.3]: https://github.com/gbrlsnchs/patricia/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/gbrlsnchs/patricia/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/gbrlsnchs/patricia/compare/v0.1.0...v0.1.1
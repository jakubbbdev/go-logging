# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of Go Logging Library
- Core logging functionality with multiple log levels
- Structured logging with fields support
- Context-aware logging
- Multiple output handlers (Console, File, Multi)
- Multiple formatters (Text, JSON)
- Thread-safe operations
- Comprehensive test suite
- Performance benchmarks
- Example applications
- Complete API documentation

### Features
- **Log Levels**: Debug, Info, Warn, Error, Fatal, Panic
- **Structured Logging**: Support for key-value fields
- **Context Support**: Automatic field inclusion from context
- **Multiple Handlers**: Console, file, and custom handlers
- **Multiple Formatters**: Text and JSON output formats
- **Color Support**: Colored output for better readability
- **Thread Safety**: Safe for concurrent use
- **Performance Optimized**: Zero-allocation logging for common cases

### Technical Details
- Go 1.21+ compatibility
- MIT License
- Comprehensive test coverage
- Benchmark suite for performance testing
- Example applications demonstrating usage
- Full API documentation 
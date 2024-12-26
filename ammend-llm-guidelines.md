# LLM Code Editing Guidelines

## Overview

This document provides guidelines for LLMs when editing existing code in the project. These guidelines complement the code generation guidelines and ensure consistent, high-quality code modifications.

## Core Principles

### Code Modification Approach

- Always preserve existing functionality unless explicitly instructed otherwise
- Maintain consistent code style with surrounding code
- Keep changes minimal and focused
- Ensure backward compatibility
- Preserve existing error handling patterns
- Maintain or improve test coverage

### Before Making Changes

1. Code Analysis
   - Thoroughly understand the existing implementation
   - Identify all affected components and dependencies
   - Review existing tests and documentation
   - Consider impact on API contracts

2. Change Planning
   - Plan changes to minimize code disruption
   - Consider breaking changes and versioning
   - Identify required documentation updates
   - Plan test modifications

## Specific Guidelines

### Code Modification Standards

1. Style Consistency
   - Match existing indentation and formatting
   - Follow established naming conventions
   - Preserve comment style and documentation format
   - Maintain consistent error handling patterns

2. Error Handling
   - Preserve existing error types and wrapping patterns
   - Maintain or enhance logging context
   - Keep consistent HTTP status code usage
   - Update error documentation if changed

3. Documentation Updates
   - Update affected documentation inline
   - Modify API documentation if endpoints change
   - Update examples to reflect changes
   - Add migration notes for breaking changes

### Testing Requirements

1. Test Modifications
   - Update affected tests to match new functionality
   - Preserve existing test patterns and style
   - Maintain or improve test coverage
   - Add tests for new edge cases

2. Test Coverage
   - Ensure modified code is fully tested
   - Update mocks as needed
   - Test both success and error paths
   - Verify backward compatibility

### Security Considerations

1. Code Review
   - Verify input validation remains intact
   - Check authorization logic is preserved
   - Ensure rate limiting remains effective
   - Maintain security best practices

2. Vulnerability Prevention
   - Review modified code for security implications
   - Update security documentation if needed
   - Maintain existing security measures
   - Add security notes for significant changes

## Implementation Examples

### Modifying Function Example

```go
// Original function
func processData(data string) (*Result, error) {
    // Existing implementation
}

// Modified function with new parameter
func processData(data string, options *Options) (*Result, error) {
    // Validate new parameter
    if options != nil {
        if err := validateOptions(options); err != nil {
            return nil, fmt.Errorf("invalid options: %w", err)
        }
    }

    // Preserve existing logic
    // Add new functionality
    // Maintain error handling pattern
}
```

### Updating Tests Example

```go
// Original test
func TestProcessData(t *testing.T) {
    // Existing test cases
}

// Modified test with new cases
func TestProcessData(t *testing.T) {
    // Preserve existing test cases
    tests := []struct {
        name    string
        data    string
        options *Options
        want    *Result
        wantErr bool
    }{
        // Original test cases preserved
        {
            name: "original case",
            data: "test",
            want: &Result{...},
        },
        // New test cases added
        {
            name: "with options",
            data: "test",
            options: &Options{...},
            want: &Result{...},
        },
    }
    // Test implementation
}
```

## Best Practices

### Code Organization

- Keep modifications within existing package boundaries
- Maintain clear separation of concerns
- Preserve or improve code organization
- Follow established project structure

### Configuration Changes

- Update configuration in @config/config.yaml
- Document new configuration options
- Provide migration guidance for config changes
- Maintain backward compatibility

### Documentation Maintenance

- Update affected documentation immediately
- Keep documentation in sync with code changes
- Provide clear upgrade instructions
- Document breaking changes prominently

## Quality Checklist

Before completing modifications:

- [ ] Code follows existing patterns and style
- [ ] All tests pass and coverage maintained
- [ ] Documentation updated
- [ ] Error handling preserved or improved
- [ ] Security measures intact
- [ ] Breaking changes documented
- [ ] Configuration updates noted
- [ ] API documentation current

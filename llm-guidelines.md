# LLM Development Guidelines

## Overview

This document outlines the guidelines and best practices for LLMs contributing to our development process. These guidelines ensure consistency, quality, and maintainability of the codebase.

## Core Principles

### Code Generation

- Generate complete, working code snippets that follow established project structure
- Include comprehensive error handling and logging
- Add detailed comments explaining complex logic and design decisions
- Follow idiomatic Go practices and conventions
- Maintain consistent formatting and style
- Never skip error handling or use placeholder comments

### API Design

- Design RESTful endpoints with clear, consistent naming
- Use appropriate HTTP methods for operations
- Implement proper request validation and error responses
- Provide comprehensive endpoint documentation
- Include example requests and responses

### Documentation

- Write clear, concise documentation for all new features
- Update existing documentation when modifying functionality
- Include practical examples and use cases
- Document configuration options and environment variables
- Provide troubleshooting guidance

### Testing Guidelines

- Write comprehensive unit tests for all business logic
- Include edge cases and error scenarios
- Do not write tests for Docker-related functionality
- Mock external dependencies appropriately
- Maintain high test coverage for non-Docker components

## Specific Guidelines

### Adding New Features

1. Requirement Analysis

   - Understand the feature requirements completely
   - Consider edge cases and failure scenarios
   - Plan for backward compatibility
   - Consider performance implications

2. Implementation

   - Follow the project's directory structure
   - Implement proper validation and error handling
   - Add comprehensive logging
   - Include necessary documentation
   - Write unit tests for business logic

3. Documentation
   - Update API documentation
   - Add usage examples
   - Document configuration options
   - Include troubleshooting information

### Modifying Existing Code

1. Analysis

   - Understand current implementation thoroughly
   - Identify potential impact areas
   - Consider backward compatibility
   - Review existing tests

2. Implementation
   - Maintain consistent code style
   - Update affected documentation
   - Modify relevant tests
   - Preserve existing error handling patterns

### Code Quality Standards

1. Error Handling

   - Use custom error types when appropriate
   - Implement proper error wrapping
   - Log errors with context
   - Return appropriate HTTP status codes

2. Comments and Documentation

   - Add package-level documentation
   - Document complex logic
   - Include examples in documentation
   - Explain non-obvious design decisions

3. Testing
   - Write tests for happy and error paths
   - Use table-driven tests when appropriate
   - Mock external dependencies
   - Focus on business logic testing

### Security Considerations

1. Input Validation

   - Validate all user inputs
   - Implement proper sanitization
   - Check parameter types and limits
   - Prevent injection attacks

2. Authentication/Authorization
   - Implement proper token validation
   - Add rate limiting where appropriate
   - Follow security best practices
   - Document security features

## Implementation Examples

### Error Handling Example

```go
func processRequest(req *Request) (*Response, error) {
    if err := validateRequest(req); err != nil {
        return nil, fmt.Errorf("invalid request: %w", err)
    }

    result, err := processData(req.Data)
    if err != nil {
        return nil, fmt.Errorf("failed to process data: %w", err)
    }

    return &Response{
        Result: result,
        Status: "success",
    }, nil
}
```

### Test Example

```go
func TestProcessRequest(t *testing.T) {
    tests := []struct {
        name    string
        req     *Request
        want    *Response
        wantErr bool
    }{
        {
            name: "valid request",
            req: &Request{
                Data: "valid data",
            },
            want: &Response{
                Result: "processed data",
                Status: "success",
            },
            wantErr: false,
        },
        {
            name: "invalid request",
            req: &Request{
                Data: "",
            },
            want:    nil,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := processRequest(tt.req)
            if (err != nil) != tt.wantErr {
                t.Errorf("processRequest() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("processRequest() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Best Practices

### Code Organization

- Keep packages focused and cohesive
- Use clear and meaningful names
- Maintain clear dependency boundaries
- Design clean interfaces

### Configuration

- Make sure to keep all the configuration in one place which is @config/config.yaml

### Documentation

- Keep documentation up to date
- Include practical examples
- Document configuration options
- Provide troubleshooting guides

### Performance

- Consider resource usage
- Implement appropriate caching
- Use efficient algorithms
- Add proper logging for debugging

Remember that these guidelines are meant to ensure consistency and quality while making the codebase maintainable and reliable. When in doubt, please ask questions!

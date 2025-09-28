# Generated Mocks

This directory contains testify mocks for all port interfaces, generated using [mockery](https://github.com/vektra/mockery).

## Available Mocks

- **`Logger.go`** - Mock for `ports.Logger` interface
- **`MessageRepository.go`** - Mock for `ports.MessageRepository` interface
- **`MessagePublisher.go`** - Mock for `ports.MessagePublisher` interface

## Usage in Tests

```go
import (
    "messaging-app/internal/mocks"
    "github.com/stretchr/testify/mock"
)

func TestMyHandler(t *testing.T) {
    // Create mock instances
    mockRepo := &mocks.MessageRepository{}
    mockPublisher := &mocks.MessagePublisher{}
    mockLogger := &mocks.Logger{}

    // Set up expectations
    mockRepo.On("SaveMessage", mock.Anything, mock.Anything).Return(nil)
    mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

    // Use mocks in your handler
    handler := NewMessageHandler(mockRepo, mockPublisher, mockLogger)

    // Execute test...

    // Verify expectations were met
    mockRepo.AssertExpectations(t)
    mockPublisher.AssertExpectations(t)
    mockLogger.AssertExpectations(t)
}
```

## Regenerating Mocks

To regenerate mocks after interface changes:

```bash
go generate ./internal/ports/...
```

## Notes

- Mocks are automatically generated from the port interfaces
- Don't edit mock files manually - they will be overwritten
- Each mock implements the full interface with testify/mock functionality
- Use `mock.Anything` for flexible argument matching
- Use specific values when you need exact matching
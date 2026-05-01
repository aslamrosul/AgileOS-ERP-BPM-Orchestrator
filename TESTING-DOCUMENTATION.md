# Automated Testing & Quality Assurance Documentation

## Overview

The AgileOS BPM platform includes a comprehensive automated testing framework that ensures code quality, stability, and reliability before deployment. The testing suite covers backend Go services, frontend Next.js components, Python analytics, and end-to-end integration scenarios.

## Testing Architecture

```
┌─────────────────────────────────────────────────────────┐
│                  Testing Framework                       │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │   Backend    │  │   Frontend   │  │   Python     │ │
│  │   Go Tests   │  │  Next.js     │  │  Analytics   │ │
│  │              │  │   Tests      │  │    Tests     │ │
│  └──────────────┘  └──────────────┘  └──────────────┘ │
│                                                          │
│  ┌──────────────────────────────────────────────────┐  │
│  │         Integration & E2E Tests                  │  │
│  └──────────────────────────────────────────────────┘  │
│                                                          │
│  ┌──────────────────────────────────────────────────┐  │
│  │         CI/CD Pipeline (GitHub Actions)          │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

## Test Coverage

### Backend Go Tests

#### Unit Tests
- **BPM Engine Tests** (`internal/bpm/engine_test.go`)
  - Workflow validation
  - Step navigation (GetNextStep)
  - Process initiation
  - Database mocking

- **Auth Middleware Tests** (`middleware/auth_test.go`)
  - JWT token validation
  - Expired token handling
  - Invalid token format detection
  - Role-based authorization
  - Multiple role scenarios

#### Test Coverage Target
- **Minimum**: 70%
- **Target**: 80%+
- **Current**: Measured automatically in CI/CD

### Frontend Next.js Tests

#### Component Tests
- **SignatureBadge Component** (`__tests__/components/SignatureBadge.test.tsx`)
  - QR code rendering
  - Signature display
  - Null/undefined handling
  - Accessibility compliance
  - Props updates

#### Testing Tools
- **Vitest**: Fast unit test framework
- **React Testing Library**: Component testing
- **jsdom**: DOM simulation

### Python Analytics Tests

#### Unit Tests
- Data processing functions
- ML model predictions
- Anomaly detection algorithms
- Statistical calculations

### Integration Tests

#### E2E Scenarios (`tests/integration_test.go`)
1. **Complete Workflow Test**
   - User login
   - Workflow creation
   - Process initiation
   - Audit trail verification

2. **Security Tests**
   - Unauthorized access handling
   - Invalid token rejection
   - Role-based access control

3. **Validation Tests**
   - Invalid data handling
   - Error responses
   - Data integrity

## Running Tests

### Quick Start

#### All Tests (Recommended)
```bash
# Linux/Mac
./test.sh

# Windows
.\test.ps1
```

### Individual Test Suites

#### Backend Go Tests
```bash
cd agile-os/backend-go

# Run all tests
go test ./... -v

# Run with coverage
go test ./... -v -cover -coverprofile=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Run specific package
go test ./internal/bpm/... -v
go test ./middleware/... -v
go test ./tests/... -v

# Run with race detection
go test ./... -race

# Benchmark tests
go test ./... -bench=. -benchmem
```

#### Frontend Next.js Tests
```bash
cd agile-os/frontend-next

# Install dependencies
npm install

# Run tests
npm run test

# Run tests with UI
npm run test:ui

# Run with coverage
npm run test:coverage

# Watch mode (development)
npm run test -- --watch
```

#### Python Analytics Tests
```bash
cd agile-os/analytics-py

# Install dependencies
pip install -r requirements.txt
pip install pytest pytest-cov

# Run tests
pytest tests/ -v

# Run with coverage
pytest tests/ -v --cov=. --cov-report=html
```

## Test Structure

### Backend Go Test Structure

```go
// Test function naming convention
func TestFunctionName_Scenario(t *testing.T) {
    // Arrange
    mockDB := new(MockDatabase)
    engine := NewBPMEngine(mockDB)
    
    // Act
    result, err := engine.GetNextStep("step_1")
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, "expected_value", result.ID)
}

// Table-driven tests
func TestValidation(t *testing.T) {
    tests := []struct {
        name      string
        input     *Workflow
        expectErr bool
    }{
        {"Valid workflow", validWorkflow, false},
        {"Invalid workflow", invalidWorkflow, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateWorkflow(tt.input)
            if tt.expectErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Frontend Test Structure

```typescript
describe('Component Name', () => {
  it('should render correctly', () => {
    render(<Component prop="value" />);
    expect(screen.getByText('Expected Text')).toBeInTheDocument();
  });

  it('should handle user interaction', async () => {
    const user = userEvent.setup();
    render(<Component />);
    
    await user.click(screen.getByRole('button'));
    expect(screen.getByText('Result')).toBeInTheDocument();
  });
});
```

## Mocking Strategies

### Database Mocking (Go)

```go
type MockDatabase struct {
    mock.Mock
}

func (m *MockDatabase) Query(query string, params map[string]interface{}) (interface{}, error) {
    args := m.Called(query, params)
    return args.Get(0), args.Error(1)
}

// Usage in tests
mockDB := new(MockDatabase)
mockDB.On("GetWorkflow", "workflow_1").Return(workflow, nil)
```

### API Mocking (Frontend)

```typescript
vi.mock('@/lib/api', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
  },
}));

// Usage in tests
(api.get as any).mockResolvedValue({ data: mockData });
```

## CI/CD Integration

### GitHub Actions Workflow

The automated testing pipeline runs on:
- **Push** to main or develop branches
- **Pull requests** to main or develop branches

#### Pipeline Stages

1. **Backend Tests**
   - Go unit tests
   - Coverage report generation
   - Coverage threshold check (70% minimum)

2. **Frontend Tests**
   - Component tests
   - Coverage report generation

3. **Integration Tests**
   - E2E scenario testing
   - Service integration validation

4. **Python Tests**
   - Analytics function tests
   - ML model validation

5. **Test Summary**
   - Aggregate results
   - Fail if any critical tests fail

### Viewing CI Results

1. Navigate to GitHub repository
2. Click "Actions" tab
3. Select workflow run
4. View test results and coverage reports

## Coverage Reports

### Backend Coverage

```bash
# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# View in browser
open coverage.html  # Mac
xdg-open coverage.html  # Linux
start coverage.html  # Windows
```

### Frontend Coverage

```bash
# Generate coverage
npm run test:coverage

# View report
open coverage/index.html
```

### Coverage Metrics

- **Line Coverage**: Percentage of code lines executed
- **Branch Coverage**: Percentage of conditional branches tested
- **Function Coverage**: Percentage of functions called
- **Statement Coverage**: Percentage of statements executed

## Best Practices

### Writing Tests

1. **Follow AAA Pattern**
   - Arrange: Set up test data
   - Act: Execute the function
   - Assert: Verify the results

2. **Test One Thing**
   - Each test should verify one specific behavior
   - Use descriptive test names

3. **Use Table-Driven Tests**
   - Test multiple scenarios efficiently
   - Improve test maintainability

4. **Mock External Dependencies**
   - Database calls
   - API requests
   - File system operations

5. **Test Edge Cases**
   - Null/undefined values
   - Empty arrays/strings
   - Boundary conditions
   - Error scenarios

### Test Naming Conventions

#### Go Tests
```go
// Format: TestFunctionName_Scenario_ExpectedBehavior
TestGetNextStep_ValidStep_ReturnsNextStep
TestGetNextStep_InvalidStep_ReturnsError
TestAuthMiddleware_ExpiredToken_ReturnsUnauthorized
```

#### Frontend Tests
```typescript
// Format: should + action + expected result
it('should render QR code when signature is provided')
it('should display error message when data is invalid')
it('should call API when button is clicked')
```

## Debugging Tests

### Go Tests

```bash
# Run specific test
go test -run TestFunctionName

# Verbose output
go test -v

# Show test coverage for specific package
go test -cover ./internal/bpm/

# Run with debugger
dlv test ./internal/bpm/
```

### Frontend Tests

```bash
# Run specific test file
npm run test SignatureBadge.test.tsx

# Debug mode
npm run test:ui

# Watch mode for development
npm run test -- --watch
```

## Performance Testing

### Benchmark Tests (Go)

```go
func BenchmarkGetNextStep(b *testing.B) {
    // Setup
    mockDB := new(MockDatabase)
    engine := NewBPMEngine(mockDB)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        engine.GetNextStep("step_1")
    }
}

// Run benchmarks
go test -bench=. -benchmem
```

## Continuous Improvement

### Adding New Tests

1. **Identify Critical Paths**
   - Business logic functions
   - Security-sensitive code
   - Data validation

2. **Write Tests First (TDD)**
   - Define expected behavior
   - Write failing test
   - Implement functionality
   - Verify test passes

3. **Maintain Coverage**
   - Add tests for new features
   - Update tests when refactoring
   - Remove obsolete tests

### Code Review Checklist

- [ ] All new code has tests
- [ ] Tests are meaningful and not trivial
- [ ] Edge cases are covered
- [ ] Mocks are used appropriately
- [ ] Tests are independent
- [ ] Coverage threshold is maintained

## Troubleshooting

### Common Issues

#### Go Tests Fail to Run
```bash
# Clean module cache
go clean -modcache

# Re-download dependencies
go mod download
go mod tidy
```

#### Frontend Tests Fail
```bash
# Clear node modules
rm -rf node_modules package-lock.json

# Reinstall
npm install
```

#### Coverage Report Not Generated
```bash
# Ensure coverage tools are installed
go install golang.org/x/tools/cmd/cover@latest
```

## Test Metrics

### Quality Metrics

- **Test Coverage**: >80% target
- **Test Execution Time**: <5 minutes for full suite
- **Test Reliability**: >99% pass rate
- **Code Quality**: No critical issues in tests

### Monitoring

- Track test execution time trends
- Monitor coverage changes over time
- Identify flaky tests
- Review test failure patterns

## Conclusion

The automated testing framework ensures that the AgileOS BPM platform maintains high quality standards throughout development. By running tests automatically on every commit and pull request, we catch bugs early and maintain confidence in our deployments.

**Remember**: Tests are not just about finding bugs—they're about documenting expected behavior and enabling confident refactoring.
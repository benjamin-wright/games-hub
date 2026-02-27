# Testing Standards

## End To End Tests
- Should test all user workflows
- Should use the same interface as the user where-ever possible

## Integration Tests
- Should test at the deployable application level
- Should test against real services where possible (e.g. an actual database)
- Should aim for the majority of tests here, save unit tests for complex logic that can't be efficiently covered at the integration level

## Unit Tests
- used for complex logical components with a high number of possible permutations
- used where there are minimal external dependencies, to promote minimal stubbing / mocking

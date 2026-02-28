# Testing Standards

## General
- Each application should have integration tests covering its core workflows.
- For anything more complex than unit tests, prefer deploying the application into a dedicated test namespace and testing against the live service.

## End To End Tests
- Should test all user workflows
- Should use the same interface as the user where-ever possible

## Integration Tests
- Should test at the deployable application level
- Should test against real services where possible (e.g. an actual database, the local test cluster, etc.)
- Should aim for the majority of tests here, save unit tests for complex logic that can't be efficiently covered at the integration level
- Should make use of shared resources where possible, (e.g. creating a single database instance to test multiple properties, rather than a separate database instance per test clause)
- Should simplify test harnesses where possible by deploying the application into an integration testing namespace (specific to the test suite) and testing against the deployed application.

## Unit Tests
- used for complex logical components with a high number of possible permutations
- used where there are minimal external dependencies, to promote minimal stubbing / mocking

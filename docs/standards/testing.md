# Testing Standards

**Governing principle:** Test at the highest level that exercises the code path efficiently. Drop to a lower level only when combinatorial complexity makes the higher level impractical.

## End-to-End Tests
- Test all user workflows through the same interface the user would use.

## Integration Tests
- Deploy the application into a dedicated test namespace and test against real services (database, cluster, etc.).
- Aim for the majority of test coverage here — prefer shared resources (e.g. one database instance for multiple assertions) over per-test isolation.

## Unit Tests
- Reserve for complex logic with many input permutations and minimal external dependencies.

## Test Design Rules

- Test through exported entry points. Never export a function solely for testability — exercise it indirectly via the public API.
- Every test double must be exercised. An unused fake indicates a missing code path — delete it or rewrite the test.
- Use `gomega` (`Expect(...).To(...)` with `RegisterTestingT(t)`) for assertions. No raw `if … { t.Errorf }`.
- If a component can't be unit-tested without its external dependency, refactor the dependency behind an interface that a fake can replace.

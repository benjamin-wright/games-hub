# General Standards

## Documentation
- Every deployable component should include a `spec.md` in its root, which should detail the features and interfaces that component provides. Specs must not restate project-wide standards (language choice, code-organisation patterns, testing practices) or describe internal implementation details — only observable behaviour and external contracts.

## Reuse Over Reinvention

Before writing anything new — utility, pattern, convention, or routine — check whether an equivalent already exists in the project, a sibling module, or an existing dependency. If it does, use it. If it does not, create it in the appropriate shared location (`tools/`) so others can reuse it.

- Never duplicate a helper inline across files; parameterise a shared function to cover variant contexts.
- Follow the conventions established in sibling applications (libraries, structure, naming). Consistency takes priority over local preference.
- Prefer library-provided functions over hand-rolled logic for sanitisation, encoding, serialisation, etc.

## Code Clarity

- Names (functions, variables, types) must be descriptive enough to make their purpose obvious without a comment.
- Comments must add information the code cannot express — explain *why*, not *what*. Never write a comment that just restates the line it sits next to.
- Prefer fewer, meaningful comments over many redundant ones.
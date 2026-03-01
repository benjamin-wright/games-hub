# General Standards

## Documentation
- Every deployable component should include a `spec.md` in its root, which should detail the features that component provides

## No Code Duplication

Before writing any new utility, helper, or shared construct, check whether an equivalent already exists in the project. If a utility you need does not exist, create it in the appropriate `tools/` location and load/import it from there. Never define the same helper inline in more than one file.

## Prefer Library Functions Over Hand-Rolled Logic

Before writing any sanitisation, escaping, encoding, or serialisation routine, check whether the dependency already provides one. If it does, use it — never reimplement it.

## Code Clarity

- Names (functions, variables, types) must be descriptive enough to make their purpose obvious without a comment.
- Comments must add information the code cannot express — explain *why*, not *what*. Never write a comment that just restates the line it sits next to.
- Prefer fewer, meaningful comments over many redundant ones.
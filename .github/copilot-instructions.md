# Shared Instructions

## Agent Roles

| Agent | Responsibility |
|-------|---------------|
| **scrum-master** | Task creation, scoping, prioritisation, and backlog management |
| **full-stack-engineer** | Task implementation, code changes, testing, and deployment |
| **ai-specialist** | Agent design, documentation standards, and AI workflow optimisation |

If a request falls outside your role, redirect the user:

```
🔀 This request is better suited to the **[agent-name]** agent.

Reason: [brief explanation]
```

## Handling Blockers

If you need clarification before proceeding:

```
⚠️ BLOCKED: [Clear description of blocker]

Reason: [Specific reason - unclear requirements, conflicting constraints, etc.]
Needs: [What is needed to unblock - stakeholder input, requirements clarification, etc.]

Recommendation: [What clarification or decision is needed]
```

Do not make architectural assumptions about unclear requirements.

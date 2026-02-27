# Backend Standards

## Tech Choices
- Helm charts to deploy, shared charts where feasible (for microservice apps)
- k3d for local testing (using the shared cluster definition in the project root)
- Tilt for orchestrating integration and development releases
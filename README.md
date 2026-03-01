# Games Hub

A technology test bed for building simple, resource-efficient games and utilities.

## Docs

- [Specifications](./docs/specifications.md): The desired features and structure of the project
- [Standards](./docs/standards): The technical standards that should be applied while making any changes
- [Tasks](./docs/tasks.md): An itemised list of the next most important changes to work make
- [Apps](./apps/[area]/[component]): Directory structure for deployable applications
- [Tools](./tools): shared config and libraries
  - [Docker](./tools/docker): Shared docker images
  - [Go](./tools/go/[module]): Go library modules
 
## Todo
- db-operator local deployment replaces the root tiltfile deployment, it should deploy its own namespaced deployment to avoid disabling the main deployment.
- refactor and add some backend
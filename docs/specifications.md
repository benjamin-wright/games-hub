# Specifications

## Platform Components

| Component | Description |
| --- | --- |
| [Makefile](../Makefile) | Commands for platform-wide features, such as starting a local kubernetes development cluster |
| [db-operator](../apps/platform/db-operator) | A custom Kubernetes operator for managing small databases. |
| [db-migrations](../apps/platform/db-migrations) | A reusable framework for applying and tracking versioned SQL schema migrations. |
| [auth-server](../apps/platform/auth-server) | A basic auth server for user access control. |
| [api-gateway](../apps/platform/api-gateway) | A gateway application that accepts WebSocket connections and passes events through to a NAT-based messaging queue. |

## App Components


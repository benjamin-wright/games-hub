# API Gateway Specification

## Purpose
Accept external client connections and route events to internal services and messaging infrastructure.

## Scope
- Accept WebSocket client connections
- Validate client access tokens
- Forward client events to the NATS messaging system

## Interfaces
- HTTP endpoint for REST API traffic
- WebSocket endpoint for client event streams
- NATS — publishes client events to subjects consumed by internal services

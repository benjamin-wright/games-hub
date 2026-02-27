# DB Operator Specification

## Purpose
Provide a CRD interface for application deployments to create and manage small databases

## Scope
- CRD to represent each database type
- CRD to represent a credential for each database type
  - specify a username and secret name in the CRD (and any DB-specific permissions)
  - operator generates a randomised password and puts username and password into the named secret
  - operator defines the user inside the database with the generated credential
  - updates status fields on CRs to reflect the status of the database and state of operations
- Support for:
  - postgres
  - redis
  - NAT

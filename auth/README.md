# Auth - Authentication and Authorisation interface

Provides a high level pluggable abstraction for authentication.

## Interface

Simplify authentication with an interface that just returns true or 
false based on the current RPC context or session id. Optionally 
returns the session information for further examination.

Granular role based authorisation and control is needed at large scale 
for access management. Goes beyond just, does this person have an 
authenticated session. Should they be allowed to access the given 
resource.

Management of auth/roles should be offloaded to a service to minimise code changes 
in each individual service. Should ideally be embedded as middleware in requests handlers 
and initialised when registering a handler.

## Supported Backends

- Oauth2
- Auth service
- ?

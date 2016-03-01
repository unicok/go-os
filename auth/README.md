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

```go
// Auth handles client side validation of authentication
// The client does not actually handle authentication itself.
// This could be an oauth2 provider, openid, basic auth, etc.
type Auth interface {
	// Determine if a request with context is authorised
	// Should extract token from the context, check with
	// the authorizer and return an err if not authed.
	// Can be used for both client and server
	Authorized(ctx context.Context, req Request) (*Token, error)
	// Retrieve a token for this client, should handle refreshing
	Token() (*Token, error)
	// Lookup a token
	Introspect(ctx context.Context) (*Token, error)
	// Revoke a token
	Revoke(t *Token) error
	// Will retrieve token from the context
	FromContext(ctx context.Context) (*Token, bool)
	// Creates a context with the token which can be
	NewContext(ctx context.Context, t *Token) context.Context
	// Retrieves token from headers
	// We may get back a partial token here
	FromHeader(map[string]string) (*Token, bool)
	// Adds token to headers
	NewHeader(map[string]string, *Token) map[string]string
	// We cache policies locally from the auth server
	Start() error
	Stop() error
	// Name
	String() string
}

func NewAuth(opts ...Option) Auth {
	return newPlatform(opts...)
}
```

## Supported Backends

- Auth service (Oauth2)

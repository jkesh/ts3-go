// Package ts3 provides a Go client for TeamSpeak ServerQuery/WebQuery.
//
// It supports:
//   - TCP and SSH ServerQuery modes
//   - HTTP/HTTPS WebQuery mode (x-api-key)
//   - Session auth/selection (login/use/logout)
//   - Common server/client/channel management methods
//   - Permission and token operations
//   - Event notification registration
//
// Most high-level methods are wrappers over Exec.
// For unsupported commands, call Exec directly with raw ServerQuery syntax.
package ts3

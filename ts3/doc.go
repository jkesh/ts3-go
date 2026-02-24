// Package ts3 provides a Go client for TeamSpeak 3 ServerQuery.
//
// It supports:
//   - TCP and SSH connection modes
//   - Session auth/selection (login/use/logout)
//   - Common server/client/channel management methods
//   - Permission and token operations
//   - Event notification registration
//
// Most high-level methods are wrappers over Exec.
// For unsupported commands, call Exec directly with raw ServerQuery syntax.
package ts3

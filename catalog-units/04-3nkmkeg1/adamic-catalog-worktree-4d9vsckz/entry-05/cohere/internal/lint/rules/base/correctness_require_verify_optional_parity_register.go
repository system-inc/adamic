package base

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers base/correctness-require-verify-optional-parity.
//
// A file of its own rather than a line in a shared register.go, so two authors landing in this new
// package at once touch nothing in common.
//
// Registered but deliberately NOT enabled anywhere. The rule belongs to api-phi-health's base
// layer, which has no VerifySettings.json, and ahra's config carries no base/ keys because base is
// that project's layer rather than this one's. Where it gets enabled is a decision for Kirk.
func init() {
	rule.Register(rule.Registration{Rule: CorrectnessRequireVerifyOptionalParity})
}

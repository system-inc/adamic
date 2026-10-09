package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers no-useless-rename.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// Registered and NOT enabled. `CohereSettings.json` already carries `"no-useless-rename": "off"`, a
// decision somebody made before the rule was ported, and running EnableRule.ts would reverse it
// through the porting process rather than because anyone changed their mind. The port is here and
// proven so that turning it on later is one config line.
//
// The decoder is hand-rolled so an explicit `false` stays distinguishable from an absent key. All
// three options already default to false, so that costs nothing today and keeps the shape right if a
// default ever moves.
func init() {
	rule.Register(rule.Registration{
		Rule:   NoUselessRename,
		Decode: DecodeNoUselessRenameOptions,
	})
}

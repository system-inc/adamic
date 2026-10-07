// Package progress embeds Kirk's micro-deadlines as shipped command configuration.
// Observations still come from landed git snapshots, not this embedded plan.
package progress

import _ "embed"

//go:embed milestones.json
var Milestones []byte

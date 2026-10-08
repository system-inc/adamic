package load

// OptionDisposition is the production decision for one stricter-only site.
// Scheduled checks are obligations on lowering, never emitted or trusted checks.
type OptionDisposition struct {
	Site   OptionSite `json:"site"`
	State  string     `json:"state"`
	Kind   string     `json:"kind,omitempty"`
	Reason string     `json:"reason,omitempty"`
}

const OptionCheckScheduled = "scheduled-check"
const OptionRemainingError = "remaining-error"

// scheduleOptionSite is the sole production admission decision. Keeping the
// census and normal compilation on this path prevents their policies drifting.
func (p *Program) scheduleOptionSite(site OptionSite) OptionDisposition {
	row := OptionDisposition{Site: site, State: OptionCheckScheduled}
	switch {
	case len(site.Options) == 1 && site.Options[0] == "JSON.stringify":
		row.Kind = "json-stringify-defined"
	case len(site.Options) > 0 && site.Options[0] == "noUncheckedIndexedAccess" && (len(site.Options) == 1 || len(site.Options) == 2 && site.Options[1] == "exactOptionalPropertyTypes"):
		// checkedIndexedRead must emit presence or refuse its representation.
		row.Kind = "indexed-presence"
	case len(site.Options) == 1 && site.Options[0] == "useUnknownInCatchVariables":
		row.Kind = "caught-type"
	default:
		row.State = OptionRemainingError
		row.Reason = "no use-site guard contract supports this diagnostic: " + site.Message
		if len(site.Options) == 1 && site.Options[0] == "exactOptionalPropertyTypes" {
			row.Kind = "optional-presence"
			row.Reason = "missing own-presence representation: 5bb775ca and alias fixes through 7e7464e6; " + site.Message
		}
	}
	return row
}

func (p *Program) OptionDispositions() []OptionDisposition {
	return append([]OptionDisposition(nil), p.optionDispositions...)
}

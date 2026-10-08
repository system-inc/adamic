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
		// Catch-derived values retain their tags; lowering checks definite typed uses.
		row.Kind = "caught-type"
	case p.acceptOptionalWrite(site), p.acceptOptionalRelation(site), p.acceptOptionalLiteral(site), p.acceptOptionalView(site), p.acceptOptionalDefined(site), p.acceptOptionalNested(site):
		row.Kind = "optional-presence"
	default:
		row.State = OptionRemainingError
		if len(site.Options) == 1 {
			switch site.Options[0] {
			case "useUnknownInCatchVariables":
				row.Kind = "catch-error"
			case "exactOptionalPropertyTypes":
				row.Kind = "optional-presence"
			}
		}
		row.Reason = "no use-site guard contract supports this diagnostic: " + site.Message
	}
	return row
}

func (p *Program) OptionDispositions() []OptionDisposition {
	return append([]OptionDisposition(nil), p.optionDispositions...)
}

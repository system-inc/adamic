package adamic

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers adamic/nominal-class. No decoder: a class type either means its instances or it does not.
func init() {
	rule.Register(rule.Registration{Rule: NominalClass})
}

package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) nullishViewField(property ir.Property) string {
	value := fmt.Sprintf("adamicViewNullish(%s, %s, %s, %s, %d, %t, %t, [%s], %t, %t)", e.nominalViewReceiver(property), quote(property.Name), quote(property.View), quote(property.ViewType), property.NullishKinds, property.NullAllowed, property.UndefinedAllowed, e.values(property.ViewAllowed), property.Absent, property.Optional)
	value = e.nominalViewRead(e.program.NominalReadContracts[property.ViewTypeID], e.mapViewCertificate(property, e.nullishMemberSelection(property, value)), property.View, property.Absent)
	return e.viewIntersectionNullishRead(property, e.viewCallableNullishCertificate(property, value))
}

func (e *emitter) nullishMemberSelection(property ir.Property, value string) string {
	if property.ViewContract == 0 {
		return value
	}
	contract := e.program.ViewContracts[property.ViewContract-1]
	if contract.Kind == ir.ViewNullable {
		if contract.Element == 0 {
			return value
		}
		property.ViewContract = contract.Element
		contract = e.program.ViewContracts[contract.Element-1]
	}
	if contract.Kind != ir.ViewUnion {
		return value
	}
	if contract.Of == ir.Object {
		return fmt.Sprintf("((v) => v == null ? v : %s)(%s)", e.viewObjectUnion(property, "v"), value)
	}
	tests := []string{}
	for _, id := range contract.Members {
		member := e.program.ViewContracts[id-1]
		if member.Kind == ir.ViewUndefined || member.Kind == ir.ViewNull {
			continue
		}
		test := fmt.Sprintf("adamicLogicalKind(v) === %d", member.Of)
		allowed := []string{}
		for _, literal := range member.Allowed {
			var constant ir.Expression
			switch literal.Of {
			case ir.Number:
				constant = ir.NumberConstant{Value: literal.Number}
			case ir.Boolean:
				constant = ir.BooleanConstant{Value: literal.Boolean}
			case ir.String:
				allowed = append(allowed, "v === "+quote(literal.String))
				continue
			}
			allowed = append(allowed, "v === "+e.value(constant))
		}
		if len(allowed) != 0 {
			test += " && (" + strings.Join(allowed, " || ") + ")"
		}
		tests = append(tests, "("+test+")")
	}
	checks := ""
	for _, id := range contract.Members {
		if e.program.ViewContracts[id-1].Kind == ir.ViewMap {
			mapped := property
			mapped.ViewContract = id
			checks += "if(v != null && adamicLogicalKind(v)===6){" + e.mapViewCertificate(mapped, "v") + ";}"
		}
	}
	return fmt.Sprintf("((v) => {if(v != null && !(%s)) panic('field read failed: '+%s+'; expected '+%s+', found '+typeof v);%s return v;})(%s)", strings.Join(tests, " || "), quote(property.View+" matches no member of "+property.ViewType), quote(property.ViewType), checks, value)
}

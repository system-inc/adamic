package javascript

import "github.com/system-inc/adamic/internal/ir"

func (e *emitter) arrayCallableCertificate(property ir.Property, value string) string {
	if !ir.ArrayCallableContract(e.program, property.ViewContract) {
		return e.emitViewCallableCertificate(property, value)
	}
	expected := e.viewCallableExpected(property)
	recorded := e.viewCallableRecorded("value", property)
	return "((value)=>{const recorded=" + recorded + ";let expected=" + expected + ";if(recorded&&expected&&expected.result===254){const arity=Math.min(recorded.parameters.length,expected.parameters.length);expected={...expected,result:255,parameters:expected.parameters.slice(0,arity),parameterMasks:expected.parameterMasks.slice(0,arity)};}else if(recorded&&expected&&expected.result!==255&&recorded.result!==0&&expected.result!==0&&recorded.result!==expected.result){panic(\"field read failed: \"+" + quote(property.View) + "+\" expected \"+expected.name+\", found function with incompatible result representation\");}return " + emitViewCallableShape("value", "recorded", "expected", property.View, property.Optional || property.Absent || property.UndefinedAllowed) + ";})(" + value + ")"
}

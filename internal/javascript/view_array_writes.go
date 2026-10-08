package javascript

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) viewArraySourceCertificate(expression ir.Expression, value string) string {
	switch source := expression.(type) {
	case ir.ArrayHoles:
		return fmt.Sprintf("adamicArrayElementCertificate(%s, %d)", value, source.ElementContract)
	case ir.ArrayLiteral:
		return fmt.Sprintf("adamicArrayElementCertificate(%s, %d)", value, source.ElementContract)
	case ir.ObjectLiteral:
		return fmt.Sprintf("adamicArrayObjectCertificate(%s, %d)", value, source.ArrayWriteContract)
	}
	return value
}

func viewArrayReferenceRuntime(program *ir.Program) string {
	names := []string{"uncertified contract"}
	type field struct {
		Name     string
		Type     ir.Type
		Expected string
		Allowed  []any
	}
	fields := map[int][]field{}
	for index, contract := range program.ViewContracts {
		names = append(names, contract.Name)
		if !ir.FlatArrayRecordContract(program, ir.ViewContractID(index+1)) {
			continue
		}
		own := []field{}
		for _, property := range contract.Fields {
			child := program.ViewContracts[property.Contract-1]
			descriptor := field{Name: property.Name, Type: child.Of, Expected: child.Name, Allowed: []any{}}
			for _, literal := range child.Allowed {
				switch literal.Of {
				case ir.Number:
					descriptor.Allowed = append(descriptor.Allowed, literal.Number)
				case ir.Boolean:
					descriptor.Allowed = append(descriptor.Allowed, literal.Boolean)
				case ir.String:
					descriptor.Allowed = append(descriptor.Allowed, literal.String)
				}
			}
			own = append(own, descriptor)
		}
		fields[index+1] = own
	}
	nominal := map[int]int{}
	nullable := map[int]bool{}
	nulls := map[int]bool{}
	for index, source := range program.ViewContracts {
		if source.Of != ir.Object && source.Of != ir.Union {
			continue
		}
		c := source
		if c.Kind == ir.ViewNullable && c.Element != 0 {
			c = program.ViewContracts[c.Element-1]
		}
		if c.NominalClass != 0 {
			nominal[index+1] = c.NominalClass
			nullable[index+1] = source.Undefined
			nulls[index+1] = source.Null
		}
	}
	classes, _ := json.Marshal(nominal)
	missing, _ := json.Marshal(nullable)
	nullJSON, _ := json.Marshal(nulls)
	named, _ := json.Marshal(names)
	pairs, _ := json.Marshal(ir.ArrayRecordWritePairs(program))
	descriptors, _ := json.Marshal(fields)
	return fmt.Sprintf("const adamicArrayNominalClasses = %s;\nconst adamicArrayNominalUndefined = %s;\nconst adamicArrayNominalNull = %s;\nconst adamicArrayReferenceNames = %s;\nconst adamicArrayReferencePairs = %s;\nconst adamicArrayReferenceFields = %s;\n", classes, missing, nullJSON, named, pairs, descriptors) + viewArrayReferenceHelpers
}

const viewArrayReferenceHelpers = `const adamicArrayElementContracts = new WeakMap();
const adamicArrayObjectContracts = new WeakMap();
const adamicArrayElementCertificate = (array, contract) => { adamicArrayElementContracts.set(array, contract); return array; };
const adamicArrayObjectCertificate = (object, contract) => { adamicArrayObjectContracts.set(object, contract); return object; };
const adamicArrayReferenceFailure = (expected, found) => panic("element write failed: <array write> expected " + expected + ", found " + found);
const adamicArrayReferenceWrite = (array, value) => {
 if (adamicArrayStorage.get(array) !== 10) adamicArrayWriteCheck(array, 4);
 const target = adamicArrayElementContracts.get(array) || 0;
 if (!target) adamicArrayReferenceFailure("object", "uncertified source element contract");
 if (adamicArrayNominalClasses[target]) {
  if (!(adamicInstanceOf(value,adamicArrayNominalClasses[target]) || value === undefined && adamicArrayNominalUndefined[target] || value === null && adamicArrayNominalNull[target])) adamicArrayReferenceFailure(adamicArrayReferenceNames[target],"object without required class identity");
  return;
 }
 if (value === undefined || value === null || typeof value !== 'object' || Array.isArray(value) || value instanceof Map) adamicArrayReferenceFailure(adamicArrayReferenceNames[target], value === undefined ? 'undefined' : 'non-object');
 const source = adamicArrayObjectContracts.get(value) || 0;
 if (!source) adamicArrayReferenceFailure(adamicArrayReferenceNames[target], "uncertified incoming record contract");
 if (!adamicArrayReferencePairs.some(pair => pair[0] === source && pair[1] === target)) adamicArrayReferenceFailure(adamicArrayReferenceNames[target], adamicArrayReferenceNames[source]);
 for (const field of adamicArrayReferenceFields[source]) adamicViewField(value, field.Name, "<array write>." + field.Name, field.Type, field.Expected, field.Allowed);
};
`

package javascript

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

// UntaggedUnionRuntime consumes shared non-panicking probes and membership
// adapters. Dispatch must preserve the returned contract on subsequent reads.
func UntaggedUnionRuntime() string { return viewUntaggedUnionsRuntime }

const viewUntaggedUnionsRuntime = `
const adamicViewUntaggedUnionSelect = (snapshot, members, probe, match, expression, declared) => {
 if (snapshot.kind === 'object' && snapshot.value !== null && snapshot.value !== undefined && match !== undefined) {
  for (const member of members) {
   if (!member.contract) continue;
   let candidate = true;
   for (const tag of member.tags) {
    const slot = probe === undefined ? undefined : probe(snapshot, tag.field);
    if (slot === undefined || !slot.present || !slot.initialized ||
        !tag.allowed.some(allowed => allowed.literal && allowed.kind === slot.snapshot.kind && allowed.value === slot.snapshot.value)) {
     candidate = false; break;
    }
   }
   if (!candidate) continue;
   if (match({kind:'object', contract:member.contract}, snapshot)) return member.contract;
  }
 }
 return adamicViewMixedUnionSelect(snapshot, [], undefined, expression, declared);
};
`

// Named handoff to shared union dispatch. The wrapper evaluates its operand once
// and every later field read still uses the shared readiness/type checks.
func (e *emitter) viewUntaggedObjectUnion(property ir.Property, value string) string {
	encoded, err := json.Marshal(e.program.ViewContracts)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("((adamicUntaggedValue) => { %s; adamicUntaggedPlainSelect(adamicUntaggedValue, %s, %d, %s, %s); return adamicUntaggedValue; })(%s)", viewUntaggedPlainRuntime, string(encoded), property.ViewContract, quote(property.View), quote(property.ViewType), value)
}

const viewUntaggedPlainRuntime = `
const adamicUntaggedPlainSelect = (value, contracts, id, expression, declared) => {
 if(value===undefined && contracts[id-1]?.Undefined) return 0;
 const slot = (object,name) => {
  if(object===null || typeof object!=='object' || Array.isArray(object) || object instanceof Map) return undefined;
  const field=Object.getOwnPropertyDescriptor(object,name);
  if(field===undefined || !Object.hasOwn(field,'value') || adamicFieldReadiness.get(object)?.has(name)) return undefined;
  return field;
 };
 const matches=(value,id)=>{
  const contract=contracts[id-1];
  if(contract===undefined || (contract.Unsupported && contract.Unsupported!=='untagged object union') || contract.Nominal) return false;
  if(value===undefined && (contract.Undefined || contract.Kind===7)) return true;
  if(contract.Kind===1){
   const wanted=contract.Of===1 || contract.Of===7?'number':contract.Of===2 || contract.Of===9?'boolean':contract.Of===3?'string':'unknown';
   if(typeof value!==wanted) return false;
   return !contract.Allowed?.length || contract.Allowed.some(literal=>value===(literal.Of===1?literal.Number:literal.Of===2?literal.Boolean:literal.String));
  }
  if(contract.Kind!==2 || value===null || typeof value!=='object' || Array.isArray(value) || value instanceof Map) return false;
  if(Object.getPrototypeOf(value)!==Object.prototype && Object.getPrototypeOf(value)!==null) return false;
  const fields=contract.Fields || [];
  const tags=fields.filter(field=>!field.Optional && contracts[field.Contract-1]?.Kind===1 && contracts[field.Contract-1]?.Allowed?.length);
  if(tags.length) return tags.every(field=>{const actual=slot(value,field.Name);return actual!==undefined && matches(actual.value,field.Contract);});
  return fields.every(field=>{const actual=slot(value,field.Name);return actual===undefined ? field.Optional && !Object.hasOwn(value,field.Name) : matches(actual.value,field.Contract);});
 };
 for(const member of contracts[id-1].Members){if(matches(value,member)) return member;}
 panic('field read failed: '+expression+' matches no member of '+declared+'; expected '+declared+', found object');
};
`

// Preserve the existing shared-discriminant dispatch.
func untaggedObjectUnion(contracts []ir.ViewContract, contract ir.ViewContract) bool {
	if len(contract.Members) == 0 {
		return false
	}
	for _, field := range contract.Fields {
		if field.Optional || field.Contract <= 0 || int(field.Contract) > len(contracts) {
			continue
		}
		child := contracts[field.Contract-1]
		if child.Kind == ir.ViewScalar && len(child.Allowed) != 0 {
			return false
		}
	}
	return true
}

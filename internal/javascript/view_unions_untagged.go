package javascript

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
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
	if property.ViewType == "" && property.ViewContract > 0 {
		property.ViewType = e.program.ViewContracts[property.ViewContract-1].Name
	}

	encoded, err := json.Marshal(e.program.ViewContracts)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("((adamicUntaggedValue) => { %s; adamicUntaggedPlainSelect(adamicUntaggedValue, %s, %d, %s, %s, %t); return adamicUntaggedValue; })(%s)", viewUntaggedPlainRuntime+"; const adamicUntaggedCallableRecorded=(value)=>"+e.viewCallableRecorded("value"), string(encoded), property.ViewContract, quote(property.View), quote(property.ViewType), ir.MixedArrayContract(e.program, property.ViewContract), value)
}

const viewUntaggedPlainRuntime = `
const adamicUntaggedPlainSelect = (value, contracts, id, expression, declared, completeObjects=false) => {
 if(value===undefined && contracts[id-1]?.Undefined) return 0;
 const slot = (object,name) => {
  if(object===null || typeof object!=='object' || Array.isArray(object) || object instanceof Map) return undefined;
  const field=Object.getOwnPropertyDescriptor(object,name);
  if(field===undefined || !Object.hasOwn(field,'value') || adamicFieldReadiness.get(object)?.has(name)) return undefined;
  return field;
 };
 const active=new Map();
 const matches=(value,id,depth=0)=>{
  const reference=value!==null && typeof value==='object';
  let seen=reference ? active.get(value) : undefined;
  if(seen?.has(id)) return true;
  if(reference){if(seen===undefined){seen=new Set();active.set(value,seen);}seen.add(id);}
  try {
  if(depth>128) return false;
  const contract=contracts[id-1];
  if(contract===undefined || (contract.Unsupported && contract.Unsupported!=='untagged object union') || contract.Nominal) return false;
  if(value===undefined && (contract.Undefined || contract.Kind===7)) return true;
  if(contract.Kind===1){
   const wanted=contract.Of===1 || contract.Of===7?'number':contract.Of===2 || contract.Of===9?'boolean':contract.Of===3?'string':'unknown';
   if(typeof value!==wanted) return false;
   return !contract.Allowed?.length || contract.Allowed.some(literal=>value===(literal.Of===1?literal.Number:literal.Of===2?literal.Boolean:literal.String));
  }
  if(contract.Kind===5){
   if(!contract.Result && !contract.DiscardResult) return false;
   const parameters=(contract.Parameters || []).map(id=>contracts[id-1]?.Of || 0);
   const expected={parameters,result:contract.DiscardResult?255:contracts[contract.Result-1]?.Of || 0};
   const recorded=adamicUntaggedCallableRecorded(value);
   if(contract.ProducerCertified && !(contract.Functions || []).includes(recorded?.function)) return false;
   return adamicViewCallableSignaturesMatch(recorded,expected);
  }
  if(contract.Kind===3){
   if(!Array.isArray(value) || !contract.Element) return false;
   for(let i=0;i<value.length;i++){
    const field=Object.getOwnPropertyDescriptor(value,String(i));
    if(field!==undefined && !Object.hasOwn(field,'value')) return false;
    if(!matches(field===undefined?undefined:field.value,contract.Element,depth+1)) return false;
   }
   return true;
  }
  if(contract.Kind===4) return (contract.Members || []).some(member=>matches(value,member,depth+1));
  if(contract.Kind!==2 || value===null || typeof value!=='object' || Array.isArray(value) || value instanceof Map) return false;
  const fields=contract.Fields || [];
  const ownKind=fields.find(field=>field.Name==='kind' && !field.Optional && contracts[field.Contract-1]?.Kind===1 && [1,2,3].includes(contracts[field.Contract-1]?.Of));
  if(!completeObjects && ownKind){const actual=slot(value,'kind');return actual!==undefined && matches(actual.value,ownKind.Contract,depth+1);}
  const tags=fields.filter(field=>!field.Optional && contracts[field.Contract-1]?.Kind===1 && contracts[field.Contract-1]?.Allowed?.length);
  if(!completeObjects && tags.length) return tags.every(field=>{const actual=slot(value,field.Name);return actual!==undefined && matches(actual.value,field.Contract,depth+1);});
  return fields.every(field=>{const actual=slot(value,field.Name);return actual===undefined ? field.Optional && !Object.hasOwn(value,field.Name) : matches(actual.value,field.Contract,depth+1);});
  } finally {if(reference){seen.delete(id);if(seen.size===0) active.delete(value);}}
 };
 for(const member of contracts[id-1].Members){if(matches(value,member)) return member;}
 panic('field read failed: '+expression+' matches no member of '+declared+'; expected '+declared+', found '+(Array.isArray(value)?'array':value===null?'null':typeof value));
};
`

// Preserve the existing shared-discriminant dispatch.
func untaggedObjectUnion(contracts []ir.ViewContract, contract ir.ViewContract) bool {
	return len(contract.Members) != 0 && !ir.ViewUnionHasDiscriminant(contracts, contract)
}

func (e *emitter) untaggedCallableUnionExpected(property ir.Property, recorded, fallback string) string {
	id := property.ViewContract
	if id <= 0 || int(id) > len(e.program.ViewContracts) {
		return fallback
	}
	root := e.program.ViewContracts[id-1]
	if root.Kind != ir.ViewUnion || root.Of != ir.Closure || len(root.Members) == 0 {
		return fallback
	}
	choices := []string{}
	for _, child := range root.Members {
		contract := e.program.ViewContracts[child-1]
		if contract.Kind != ir.ViewCallable || contract.Result == 0 {
			continue
		}
		parameters := make([]ir.Type, len(contract.Parameters))
		known := true
		for i, parameter := range contract.Parameters {
			parameters[i] = e.program.ViewContracts[parameter-1].Of
			known = known && parameters[i] != 0
		}
		result := e.program.ViewContracts[contract.Result-1].Of
		if !known || result == 0 {
			continue
		}
		choices = append(choices, viewCallableSignature(parameters, result, root.Name))
	}
	return "((recorded)=>{const choices=[" + strings.Join(choices, ",") + "];return choices.find(expected=>recorded!==undefined && recorded.result===expected.result && recorded.parameters.length===expected.parameters.length && recorded.parameters.every((value,index)=>value!==0 && value===expected.parameters[index])) || choices[0];})(" + recorded + ")"
}

func (e *emitter) untaggedCallableRecorded(property ir.Property, recorded, expected string) string {
	id := property.ViewContract
	if id <= 0 || int(id) > len(e.program.ViewContracts) {
		return recorded
	}
	root := e.program.ViewContracts[id-1]
	if root.Kind != ir.ViewUnion || root.Of != ir.Closure {
		return recorded
	}
	functions := []int{}
	certified := false
	for _, member := range root.Members {
		contract := e.program.ViewContracts[member-1]
		certified = certified || contract.ProducerCertified
		functions = append(functions, contract.Functions...)
	}
	if !certified {
		return recorded
	}
	encoded, _ := json.Marshal(functions)
	return "((recorded)=>!adamicViewCallableSignaturesMatch(recorded," + expected + ") || " + string(encoded) + ".includes(recorded?.function) ? recorded : undefined)(" + recorded + ")"
}

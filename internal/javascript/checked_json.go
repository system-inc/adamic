package javascript

import (
	"encoding/json"
	"github.com/system-inc/adamic/internal/ir"
)

const checkedJSONRuntime = `
const adamicCheckJSON = (value, contract, path, terminal = true, depth = 0) => {
 const found = value === null ? 'null' : Array.isArray(value) ? 'array' : adamicTypeOf(value);
 const fail = (actual = found) => { if (terminal) panic('checked any: ' + path + ' needs ' + contract.Name + ', found ' + actual); return false; };
 if (contract.Kind === 'union') {
  for (const alternative of contract.Alternatives) if (adamicCheckJSON(value, alternative, path, false, depth)) return true;
  return fail();
 }
 const domain = contract.Kind === 'json';
 if ((domain ? !['number','boolean','string','null','undefined','object','array'].includes(found) : found !== contract.Kind.replace('_literal','')) || depth > 64) return fail(depth > 64 ? 'non-JSON recursion' : found);
 if(contract.Kind.endsWith('_literal') && value !== (found==='number'?contract.LiteralNumber:found==='boolean'?contract.LiteralBoolean:contract.LiteralText)) return fail();
 if (found === 'number' && !Number.isFinite(value)) return fail('non-JSON number');
 if (found === 'object') {
  if (Object.getPrototypeOf(value) !== Object.prototype && Object.getPrototypeOf(value) !== null) return fail('class object');
  if(domain) for(const key of Object.keys(value)) if(!adamicCheckJSON(value[key],contract,path+'.'+key,terminal,depth+1)) return false;
  for (const field of contract.Fields ?? []) {
   const child = value[field.Name];
   if (field.Optional && child === undefined) continue;
   if (!adamicCheckJSON(child, field.Contract, path + '.' + field.Name, terminal, depth + 1)) return false;
  }
 }
 if (found === 'array' && (contract.Element || domain)) {
  for (let index = 0; index < value.length; index++) if (!adamicCheckJSON(value[index], domain ? contract : contract.Element, path + '[' + index + ']', terminal, depth + 1)) return false;
 }
 return true;
};
const adamicCheckedJSON = (value, contract, path) => { adamicCheckJSON(value, {Kind:'json',Name:'JSON value | undefined'}, path); adamicCheckJSON(value, contract, path); return value; };
`

func checkedJSONSchema(c *ir.JSONContract) string {
	data, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(data)
}
func (e *emitter) checkedJSON(v ir.CheckedJSON) string {
	return "adamicCheckedJSON(" + e.value(v.Value) + ", " + checkedJSONSchema(v.Contract) + ", " + quote(v.Path) + ")"
}
func runtimeJSONType(to ir.Type) *ir.JSONContract {
	kind := map[ir.Type]string{ir.Number: "number", ir.Boolean: "boolean", ir.String: "string", ir.Array: "array", ir.Object: "object"}[to]
	if to == ir.Union {
		return &ir.JSONContract{Kind: "union", Name: "JSON value", Alternatives: []*ir.JSONContract{{Kind: "undefined", Name: "undefined"}, {Kind: "null", Name: "null"}, {Kind: "number", Name: "number"}, {Kind: "boolean", Name: "boolean"}, {Kind: "string", Name: "string"}, {Kind: "object", Name: "object"}, {Kind: "array", Name: "array"}}}
	}
	if to.IsMaybe() {
		base := runtimeJSONType(to.Present())
		return &ir.JSONContract{Kind: "union", Name: base.Name + " | undefined", Alternatives: []*ir.JSONContract{base, {Kind: "undefined", Name: "undefined"}}}
	}
	if kind == "" {
		panic("unsupported checked JSON read type")
	}
	base := &ir.JSONContract{Kind: kind, Name: kind}
	if to.IsReference() {
		return &ir.JSONContract{Kind: "union", Name: kind + " | undefined", Alternatives: []*ir.JSONContract{base, {Kind: "undefined", Name: "undefined"}}}
	}
	return base
}

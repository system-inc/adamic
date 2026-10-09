package javascript

import (
	"encoding/json"
	"github.com/system-inc/adamic/internal/ir"
)

const checkedJSONRuntime = `
const adamicCheckJSON = (value, contract, path, terminal = true, depth = 0) => {
 const found = value === null ? 'null' : Array.isArray(value) ? 'array' : adamicTypeOf(value);
 const fail = (actual = found) => { if (terminal) panic('checked any: ' + path + ' needs ' + contract.Name + ', found ' + actual); return false; };
 if (depth > 64) return fail('non-JSON recursion');
 if (contract.Kind === 'ref') return adamicCheckJSON(value,contract.Element,path,terminal,depth);
 if (contract.Kind === 'union') {
  for (const alternative of contract.Alternatives) if (adamicCheckJSON(value, alternative, path, false, depth)) return true;
  if(terminal) {
   const matches=(c,d=0)=>d<=64 && (c.Kind==='ref'?matches(c.Element,d+1):c.Kind==='union'?c.Alternatives.some(a=>matches(a,d+1)):c.Kind.replace('_literal','')===found);
   const candidates=contract.Alternatives.filter(c=>matches(c));
   if(candidates.length===1) return adamicCheckJSON(value,candidates[0],path,true,depth);
  }
  return fail();
 }
 const domain = contract.Kind === 'json';
 if ((domain ? !['number','boolean','string','null','undefined','object','array'].includes(found) : found !== contract.Kind.replace('_literal','')) || depth > 64) return fail(depth > 64 ? 'non-JSON recursion' : found);
 if(contract.Kind.endsWith('_literal') && value !== (found==='number'?contract.LiteralNumber:found==='boolean'?contract.LiteralBoolean:contract.LiteralText)) return fail();
 if (found === 'number' && !Number.isFinite(value)) return fail('non-JSON number');
 if (found === 'object') {
  if (Object.getPrototypeOf(value) !== Object.prototype && Object.getPrototypeOf(value) !== null) return fail('class object');
  if(domain || contract.Element) for(const key of Object.keys(value)) if(!adamicCheckJSON(value[key],domain ? contract : contract.Element,path+'.'+key,terminal,depth+1)) return false;
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

func checkedJSONSchema(root *ir.JSONContract) string {
	nodes := ir.JSONContractGraph(root)
	definitions := map[int]*ir.JSONContract{}
	for _, c := range nodes {
		if c.Kind != "ref" {
			definitions[c.ID] = c
		}
	}
	ids := map[*ir.JSONContract]int{}
	for i, c := range nodes {
		ids[c] = i
	}
	var wire []map[string]any
	for _, c := range nodes {
		item := map[string]any{"Kind": c.Kind, "Name": c.Name, "LiteralText": c.LiteralText, "LiteralNumber": c.LiteralNumber, "LiteralBoolean": c.LiteralBoolean}
		if c.Kind == "ref" {
			item["ElementID"] = ids[definitions[c.Reference]]
		}
		if c.Element != nil {
			item["ElementID"] = ids[c.Element]
		}
		var fields []map[string]any
		for _, f := range c.Fields {
			fields = append(fields, map[string]any{"Name": f.Name, "Optional": f.Optional, "ContractID": ids[f.Contract]})
		}
		item["Fields"] = fields
		var alternatives []int
		for _, a := range c.Alternatives {
			alternatives = append(alternatives, ids[a])
		}
		item["AlternativeIDs"] = alternatives
		wire = append(wire, item)
	}
	data, err := json.Marshal(wire)
	if err != nil {
		panic(err)
	}
	return "(()=>{const nodes=" + string(data) + ";for(const node of nodes){if(node.ElementID!==undefined)node.Element=nodes[node.ElementID];for(const field of node.Fields??[])field.Contract=nodes[field.ContractID];node.Alternatives=(node.AlternativeIDs??[]).map(id=>nodes[id]);}return nodes[0];})()"
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

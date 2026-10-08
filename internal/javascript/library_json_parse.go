package javascript

import (
	"strconv"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) jsonParse(p ir.JSONParse) string {
	return "adamicJSONParse(" + e.value(p.Text) + ", " + jsonParseSchema(p.Check) + ")"
}

// Use the source string encoder: schemas can contain lone UTF-16 surrogates.
func jsonParseSchema(s *ir.JSONParseSchema) string {
	if s == nil {
		return "null"
	}
	fields := make([]string, len(s.Fields))
	for i, f := range s.Fields {
		fields[i] = "{Name:" + quote(f.Name) + ",Schema:" + jsonParseSchema(f.Schema) + "}"
	}
	members := make([]string, len(s.Members))
	for i, m := range s.Members {
		members[i] = jsonParseSchema(m)
	}
	return "{Kind:" + quote(s.Kind) + ",Name:" + quote(s.Name) + ",Optional:" + strconv.FormatBool(s.Optional) + ",HasLiteral:" + strconv.FormatBool(s.HasLiteral) + ",Literal:" + quote(s.Literal) + ",Element:" + jsonParseSchema(s.Element) + ",Fields:[" + strings.Join(fields, ",") + "],Members:[" + strings.Join(members, ",") + "]}"
}

const jsonParseRuntime = `const adamicJSONParse = (text, schema) => {
 const value = JSON.parse(text);
 const kind = x => x === null ? 'null' : Array.isArray(x) ? 'array' : typeof x;
 const matches = (x, s, present) => {
  if (s.Kind === 'raw') return true;
  if (!present) return s.Optional || s.Kind === 'undefined';
  if (s.Kind === 'union') return s.Members.some(m => matches(x, m, present));
  if (kind(x) !== s.Kind) return false;
  if (s.HasLiteral) return x === (s.Kind === 'number' ? Number(s.Literal) : s.Kind === 'boolean' ? s.Literal === 'true' : s.Literal);
  return true;
 };
 const validate = (x, s, path, present = true) => {
  if (!matches(x, s, present)) panic('boundary check: ' + path + ' expected ' + s.Name + ', got ' + (present ? kind(x) : 'undefined'));
  if (!present || s.Kind === 'raw') return;
  if (s.Kind === 'union') {validate(x, s.Members.find(m => matches(x, m, present)), path, present);return;}
  if (s.Kind === 'object') for (const field of s.Fields || []) validate(x[field.Name], field.Schema, path + '.' + field.Name, Object.hasOwn(x, field.Name));
  if (s.Kind === 'array') for (let i = 0; i < x.length; i++) validate(x[i], s.Element, path + '[' + i + ']');
 };
 if (schema !== null) validate(value, schema, '$');
 return schema === null ? undefined : value;
};
`

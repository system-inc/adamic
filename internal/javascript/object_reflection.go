package javascript

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

const objectReflectionRuntime = `
const adamicReflectionMatches = (value, members) => members.some(member => {
 const kind = value === undefined ? 10 : typeof value === 'number' ? 1 : typeof value === 'boolean' ? 2 : typeof value === 'string' ? 3 : 0;
 if (kind !== member.Kind) return false;
 if (!member.Literal) return true;
 return value === (kind === 1 ? member.Number : kind === 2 ? member.Boolean : member.Text);
});
const adamicCheckedEntries = (object, members, message) => {
 const result = [];
 for (const key of Object.keys(object)) {
  const value = object[key];
  if (!adamicReflectionMatches(value, members)) panic(message);
  result.push([key, value]);
 }
 return result;
};

`

func (e *emitter) checkedObjectCall(call ir.ObjectCall) string {
	proof := call.Reflection
	encode := func(value any) string {
		text, err := json.Marshal(value)
		if err != nil {
			panic(err)
		}
		return string(text)
	}
	if call.Method == "entries" {
		return fmt.Sprintf("adamicCheckedEntries(%s, %s, %s)", e.values(call.Arguments), encode(proof.Members), quote(proof.Message))
	}
	return fmt.Sprintf("Object.assign(%s)", e.values(call.Arguments))
}

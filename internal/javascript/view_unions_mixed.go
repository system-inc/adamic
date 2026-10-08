package javascript

import "github.com/system-inc/adamic/internal/ir"

// MixedUnionRuntime is included by the common dispatcher once its normalized
// field probe and complete member-contract adapter are wired. It never reads a
// field or invokes a getter while testing alternative member kinds.
func MixedUnionRuntime() string { return viewMixedUnionsRuntime }

const viewMixedUnionsRuntime = `
const adamicViewMixedUnionSelect = (snapshot, members, match, expression, declared) => {
 for (let index = 0; index < members.length; index++) {
  const member = members[index];
  if (snapshot.kind === 'unknown' || member.kind !== snapshot.kind) continue;
  if (member.literal && snapshot.value !== member.value) continue;
  if (['object', 'array', 'Map', 'function'].includes(member.kind)) {
   if (!member.contract || match === undefined || !match(member, snapshot)) continue;
  }
  return index;
 }
 const found = snapshot.kind === 'unknown' ? 'unsupported representation' : snapshot.kind;
 panic('cast failed: field read failed: ' + expression + ' matches no member of ' + declared + '; expected ' + declared + ', found ' + found);
};
`

// Required fields can contain undefined without permitting their absence.
func (e *emitter) viewStringUndefined(property ir.Property) bool {
	id := property.ViewContract
	return property.Of == ir.String && id > 0 && int(id) <= len(e.program.ViewContracts) && e.program.ViewContracts[id-1].Undefined && e.program.ViewContracts[id-1].Unsupported == ""
}

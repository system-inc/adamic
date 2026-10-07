package javascript

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

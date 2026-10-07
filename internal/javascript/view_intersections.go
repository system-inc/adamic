package javascript

// IntersectionRuntime consumes shared normalized snapshots. It grants neither
// readiness nor ownership and cannot replace the receiver's per-read guard.
func IntersectionRuntime() string { return viewIntersectionsRuntime }

const viewIntersectionsRuntime = `
const adamicViewIntersectionMatches = (snapshot, members, match) => {
 if (snapshot === undefined || snapshot.kind === 'unknown' || members.length === 0 || match === undefined) return false;
 for (const member of members) {
  if (!member || !match(member, snapshot)) return false;
 }
 return true;
};
const adamicViewIntersectionRequire = (snapshot, members, match, expression, declared) => {
 if (adamicViewIntersectionMatches(snapshot, members, match)) return;
 const found = snapshot === undefined || snapshot.kind === 'unknown' ? 'unsupported representation' : snapshot.kind;
 panic('field read failed: ' + expression + ' does not satisfy every member; expected ' + declared + ', found ' + found);
};
`

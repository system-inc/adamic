
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
 panic('field read failed: ' + expression + ' matches no member of ' + declared + '; expected ' + declared + ', found ' + found);
};

console.log(adamicViewMixedUnionSelect({kind:'number',value:42},[{kind:'number'}],undefined,'view.value','number'));

const snapshot = {kind:'number',value:42};
const members = [{kind:'number'}];
console.log(members.findIndex(member => member.kind === snapshot.kind));

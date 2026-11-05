interface Link { label: string; next: Link | undefined; }
function join(left: Link, right: Link): void { left.next = right; }
const first: Link = { label: 'first', next: undefined };
const second: Link = { label: 'second', next: undefined };
join(first, second);
first.next = undefined;
console.log(`${second.next?.label} ${first.next === undefined}`);

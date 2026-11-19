interface Seen { tag: string | undefined; readonly keep: boolean }
const stash: { tag: string }[] = [];
const seen: Seen & ThisType<{ tag: string }> = { tag: "a", get keep() { stash.push(this); return true; } };
console.log(String(seen.keep));
seen.tag = undefined;
const first = stash[0];
if (first !== undefined) console.log(String(first.tag.length));

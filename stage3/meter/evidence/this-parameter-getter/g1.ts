interface Seen { tag: string | undefined; readonly keep: boolean }
const stash: { tag: string }[] = [];
const seen: Seen & ThisType<{ tag: string }> = { tag: "a", get keep() { stash.push(this); return true; } };
console.log(String(seen.keep));
seen.tag = undefined;
console.log(String(stash[0]?.tag));

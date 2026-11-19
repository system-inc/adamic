const stash: { tag: string }[] = [];
const seen: { tag: string | undefined; keep(): void } = { tag: "a", keep(this: { tag: string }) { stash.push(this); } };
seen.keep();
seen.tag = undefined;
const first = stash[0];
if (first !== undefined) console.log(String(first.tag.length));

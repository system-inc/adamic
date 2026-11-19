const stash: { tag: string }[] = [];
const seen: { tag: string | undefined; keep(): void } = { tag: "a", keep(this: { tag: string }) { stash.push(this); } };
seen.keep();
seen.tag = undefined;
console.log(String(stash[0]!.tag.length));

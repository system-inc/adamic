const stash: { tag: string }[] = [];
const seen: { tag: string | undefined; keep(): void } = { tag: "a", keep(this: { tag: string }) { console.log(this.tag); } };
seen.keep();
seen.tag = undefined;
console.log(String(stash[0]?.tag));

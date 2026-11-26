interface Link {
    readonly entries: Map<string, Link[] | null | undefined | string>;
}
function find(link: Link, present: boolean): Link[] | null {
    if (present) return [link];
    return null;
}
function observe(holder: Link, link: Link): void {
    for (const key of ["self", "none", "missing"]) {
        console.log(holder.entries.has(key) ? "true" : "false");
        const value = holder.entries.get(key);
        if (value === null) console.log("null");
        else if (value === undefined) console.log(holder.entries.has(key) ? "stored undefined" : "absent");
        else if (typeof value === "string") console.log("other");
        else console.log(value[0] === link ? "same reference" : "other");
    }
}
function run(): void {
    const link: Link = { entries: new Map<string, Link[] | null | undefined | string>() };
    link.entries.set("self", find(link, true));
    link.entries.set("none", find(link, false));
    const second: Link = {
        entries: new Map<string, Link[] | null | undefined | string>([["self", find(link, true)], ["none", find(link, false)]]),
    };
    const back: Link[] = [second];
    link.entries.set("back", back);
    console.log("set");
    observe(link, link);
    console.log("constructor");
    observe(second, link);
}
run();

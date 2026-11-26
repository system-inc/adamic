interface Link {
    readonly entries: Map<string, Link | null | undefined | string>;
}
function find(link: Link, present: boolean): Link | null {
    return present ? link : null;
}
function run(): void {
    const link: Link = { entries: new Map<string, Link | null | undefined | string>() };
    link.entries.set("self", find(link, true));
    link.entries.set("none", find(link, false));
    for (const key of ["self", "none", "missing"]) {
        console.log(link.entries.has(key) ? "true" : "false");
        const value = link.entries.get(key);
        if (value === null) console.log("null");
        else if (value === undefined) console.log(link.entries.has(key) ? "stored undefined" : "absent");
        else console.log(value === link ? "same reference" : "other");
    }
}
run();

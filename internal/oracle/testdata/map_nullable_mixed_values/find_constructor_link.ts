interface Link {
    readonly entries: Map<string, Link | null | undefined | string>;
}
function find(link: Link, present: boolean): Link | null {
    return present ? link : null;
}
function run(): void {
    const link: Link = { entries: new Map<string, Link | null | undefined | string>() };
    const second: Link = {
        entries: new Map<string, Link | null | undefined | string>([["self", find(link, true)], ["none", find(link, false)]]),
    };
    link.entries.set("back", second);
    for (const key of ["self", "none", "missing"]) {
        console.log(second.entries.has(key) ? "true" : "false");
        const value = second.entries.get(key);
        if (value === null) console.log("null");
        else if (value === undefined) console.log(second.entries.has(key) ? "stored undefined" : "absent");
        else console.log(value === link ? "same reference" : "other");
    }
}
run();

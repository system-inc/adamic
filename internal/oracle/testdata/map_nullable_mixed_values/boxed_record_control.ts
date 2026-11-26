interface Link {
    readonly entries: Map<string, Link | Record<string, number> | null | undefined>;
}
function close(link: Link): void {
    // The slot type can hold Record<string, number>, so the first seed write stops even without the record write below.
    link.entries.set("self", link);
    const record: Record<string, number> = { k: 1 };
    link.entries.set("record", record);
}
function run(): void {
    const link: Link = { entries: new Map<string, Link | Record<string, number> | null | undefined>() };
    close(link);
    console.log(link.entries.has("record") ? "record present" : "record absent");
}
run();

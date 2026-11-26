interface Link {
    readonly entries: Map<string, Link[] | null | undefined | string>;
}
function close(link: Link): void {
    const back: Link[] = [link];
    link.entries.set("self", back);
}
function run(): void {
    const link: Link = { entries: new Map<string, Link[] | null | undefined | string>() };
    close(link);
    console.log(link.entries.has("self") ? "true" : "false");
}
run();

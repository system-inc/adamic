function add(values: string[] | undefined, more: string[]): void {
    console.log(String(values?.push(...more)));
    console.log(values?.join(",") ?? "absent");
}
add(["first"], []);
add(["first"], ["next", "last"]);

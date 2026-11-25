function add(values: string[] | undefined, more: string[]): void {
    console.log(String(values?.unshift(...more)));
    console.log(values?.join(",") ?? "absent");
}
add(["last"], ["first", "next"]);
add(undefined, ["first"]);

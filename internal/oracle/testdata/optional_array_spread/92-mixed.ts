function add(values: number[] | undefined, more: number[]): void {
    console.log(String(values?.push(1, ...more, 4)));
    console.log(values?.join(",") ?? "absent");
}
add([0], [2, 3]);
add(undefined, [2, 3]);

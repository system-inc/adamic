function join(values: string[] | undefined, separators: [string]): void {
    console.log(values?.join(...separators) ?? "absent");
}
join(["first", "next"], [","]);

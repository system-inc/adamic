function incoming(): string[] { console.log("evaluated"); return ["next"]; }
function merge(values: string[] | null | undefined): void {
    console.log(String(values?.push(...incoming())));
}
merge(null);
merge(undefined);
merge(["first"]);

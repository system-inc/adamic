interface Fields {
    readonly value?: string | number | null;
}

function show(fields: Fields | undefined): void {
    const value = fields?.value;
    if (value === null) console.log("null");
    else if (value === undefined) console.log("undefined");
    else if (typeof value === "string") console.log(value);
    else if (typeof value === "number") console.log(String(value));
    else console.log(typeof value);
}

const table: Record<string, string | number | boolean | null> = {};
const key: string = "value";
table[key] = "entry";
show({ value: "entry" });
table[key] = 7;
show({ value: 7 });
table[key] = null;
show({ value: null });
delete table[key];
show({});
show(undefined);

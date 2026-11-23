interface Fields {
    readonly value?: string | number | null;
}

function show(fields: Fields): void {
    const value = fields.value;
    if (value === null) console.log("null");
    else if (value === undefined) console.log("undefined");
    else if (typeof value === "string") console.log(value);
    else if (typeof value === "number") console.log(String(value));
    else console.log(value ? "true" : "false");
}

const table: Record<string, string | number | boolean | null> = {};
const key: string = "value";
table[key] = true;
show({ value: "entry" });

// The decoder site family 46 clears, builder.ts:2255:18 in tsc's adapted tree (toBuilderStateFileInfoForMultiEmit,
// unit :2252:1), in its shape: a field declared string | false read through a type a record may reach, in a program
// with record storage, on the decoder's three branches (a string signature, false, the field absent) and a plain string.
interface Fields {
    readonly version?: string;
    readonly signature?: string | false;
}

function decode(fileInfo: string | Fields): string {
    return typeof fileInfo === "string" ? "plain " + fileInfo :
        typeof fileInfo.signature === "string" ? "signed " + fileInfo.signature :
        fileInfo.signature === false ? "unsigned " + String(fileInfo.version) : "version " + String(fileInfo.version);
}

function show(fields: Fields): void {
    console.log(decode(fields));
}

const table: Record<string, string | number | boolean | null> = {};
const key: string = "signature";
table["version"] = "v2";
console.log(decode("v1"));
table[key] = "s2";
show({ version: "v2", signature: "s2" });
table[key] = false;
show({ version: "v2", signature: false });
delete table[key];
show({ version: "v2" });

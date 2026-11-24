interface Info { readonly version: string; signature: string | undefined }
type Encoded = Omit<Info, "signature"> & { signature: string | false | undefined };
function decode(x: Encoded): Info {
    return typeof x.signature === "string" ? x as Info : { version: x.version, signature: undefined };
}
console.log(String(decode({ version: "v", signature: "s" }).signature));
console.log(String(decode({ version: "v", signature: false }).signature));

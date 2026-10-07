// Each builder names the declaration the suffix is missing from.
export function messageNoTypeAliasSuffix(name: string): string {
    return `Type alias "${name}" should end in "Type", "Properties", "Interface", or "Options". Rename to "${name}Type", "${name}Properties" if it shapes React component props, "${name}Options" if it is an options or config bag, or convert it to an interface and use "${name}Interface".`;
}

// messageNoInterfaceSuffix names the interface.
export function messageNoInterfaceSuffix(name: string): string {
    return `Interface "${name}" should end in "Interface", "Properties", or "Options". Rename to "${name}Interface", "${name}Properties" if it shapes React component props, or "${name}Options" if it is an options or config bag.`;
}

// messageNoConstEnumSuffix names the const enum-shaped object.
export function messageNoConstEnumSuffix(name: string): string {
    return `Const enum-shaped object "${name}" should end in "Kind". Rename to "${name}Kind" with a paired "${name}KindType" alias, so the runtime value and the type that indexes it are visibly one thing.`;
}

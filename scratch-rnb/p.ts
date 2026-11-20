interface Box<T> { f: T }
function read<T extends string | undefined>(o: Box<T>): string { const v = o.f; if (v === undefined) { return 'u'; } return v; }
console.log(read<string | undefined>({ f: undefined }));
console.log(read<undefined>({ f: undefined }));
console.log(read<string>({ f: 's' }));

function pattern(text: string): RegExp | undefined { return new RegExp(text); }
console.log(String(pattern("entry")?.test("entry.js")));

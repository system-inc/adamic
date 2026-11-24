const shadow: { name: string; message: string; stack?: string; cause?: unknown; hasOwnProperty: (key: number) => boolean } = { name: "n", message: "m", stack: "s", hasOwnProperty: (_key: number): boolean => false };
const fake: Error = shadow;
console.log(String(fake.hasOwnProperty("message")));

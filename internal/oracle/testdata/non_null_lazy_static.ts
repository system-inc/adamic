function initial(): string | undefined { return undefined; }
class Frame { static text = initial()!; }
console.log(`${Frame.text}`);

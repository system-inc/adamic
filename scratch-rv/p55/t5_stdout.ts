interface Writer {
    write(text: string): boolean;
}
const writer: Writer = process.stdout;
console.log(writer.write ? "has write" : "no write");

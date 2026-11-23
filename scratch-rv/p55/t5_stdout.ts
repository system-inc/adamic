import 'node:process';
interface Writer {
    write(text: string): boolean;
}
const writer: Writer = process.stdout;
if (writer.write) {
    writer.write("has write\n");
} else {
    console.log("no write");
}

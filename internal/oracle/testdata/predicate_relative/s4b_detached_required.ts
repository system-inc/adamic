interface SourceFile { name: string }
interface Related { category: number; code: number; file: SourceFile | undefined; start: number | undefined; length: number | undefined; messageText: string }
interface Detached extends Related { file: undefined; fileName: string; start: number; length: number }
function isDetached(diagnostic: Related | Detached): diagnostic is Detached {
	return diagnostic.code > 0;
}
const items: (Related | Detached)[] = [{ category: 1, code: 2, file: undefined, start: 3, length: 4, messageText: "m", fileName: "f.ts" }, { category: 1, code: 0, file: { name: "g.ts" }, start: undefined, length: undefined, messageText: "n" }];
for (const item of items) {
	if (isDetached(item) && item.fileName === "f.ts") {
		console.log("detached " + item.fileName + " " + String(item.start + item.length));
	}
	else {
		console.log("related " + String(item.code));
	}
}

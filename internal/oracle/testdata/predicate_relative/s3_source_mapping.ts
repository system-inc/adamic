interface Mapping { generatedLine: number; generatedCharacter: number; sourceIndex?: number; sourceLine?: number; sourceCharacter?: number; nameIndex?: number }
interface SourceMapping extends Mapping { sourceIndex: number; sourceLine: number; sourceCharacter: number }
function isSourceMapping(mapping: Mapping): mapping is SourceMapping {
	return mapping.sourceIndex !== undefined && mapping.sourceLine !== undefined && mapping.sourceCharacter !== undefined;
}
function process(mapping: Mapping): string {
	let source = "none";
	if (isSourceMapping(mapping)) {
		source = String(mapping.sourceIndex) + ":" + String(mapping.sourceLine) + ":" + String(mapping.sourceCharacter);
	}
	return String(mapping.generatedLine) + "," + String(mapping.generatedCharacter) + " " + source + " " + String(mapping.sourceIndex) + " " + String(mapping.nameIndex);
}
console.log(process({ generatedLine: 1, generatedCharacter: 2 }));
console.log(process({ generatedLine: 3, generatedCharacter: 4, sourceIndex: 0, sourceLine: 5, sourceCharacter: 6, nameIndex: 7 }));
const all: Mapping[] = [{ generatedLine: 1, generatedCharacter: 1 }, { generatedLine: 2, generatedCharacter: 2, sourceIndex: 1, sourceLine: 2, sourceCharacter: 3 }];
console.log(String(all.filter(isSourceMapping).length));

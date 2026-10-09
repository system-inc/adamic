class Stack<Item> {
	private readonly items: Item[] = [];

	push(item: Item): void {
		this.items.push(item);
	}

	pop(): Item | undefined {
		return this.items.pop();
	}

	size(): number {
		return this.items.length;
	}
}

const openers = new Map<string, string>([
	[')', '('],
	[']', '['],
	['}', '{'],
]);

function isBalanced(input: string): boolean {
	const stack = new Stack<string>();
	for (const character of input) {
		if (character === '(' || character === '[' || character === '{') {
			stack.push(character);
			continue;
		}
		const opener = openers.get(character);
		if (opener !== undefined && stack.pop() !== opener) {
			return false;
		}
	}
	return stack.size() === 0;
}

for (const input of ['([]{})', '([)]', '((', '', 'f(x[0]) { }']) {
	console.log(`'${input}' ${isBalanced(input) ? 'balanced' : 'unbalanced'}`);
}

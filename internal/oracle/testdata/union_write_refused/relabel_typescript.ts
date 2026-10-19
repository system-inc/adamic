type Token = { kind: 'Word'; payload: string } | { kind: 'Count'; payload: number };
function relabel(token: Token): void {
	token.kind = 'Count';
}
function describe(token: Token): string {
	switch (token.kind) {
		case 'Word':
			return token.payload.toUpperCase();
		case 'Count':
			return `${token.payload * 2}`;
	}
}
const word: Token = { kind: 'Word', payload: 'ahra' };
relabel(word);
console.log(describe(word));

// Gap 9: panic(...) as a returned value is not lowered.
import { panic } from 'adamic';

function kindOf(kind: string): number {
	switch (kind) {
		case 'one':
			return 1;
	}
	return panic(`no kind ${kind}`);
}

console.log(`${kindOf('one')}`);

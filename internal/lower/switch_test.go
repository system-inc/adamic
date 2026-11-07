package lower

import "testing"

func TestSwitchConditionalBreakFallsThrough(test *testing.T) {
	test.Parallel()
	lowersAndAgreesWithNode(test, `function stop(): boolean { return false; }
function describe(value: number): string {
    let text = '';
    switch (value) {
        case 1:
            text += 'A';
            if (stop()) { break; }
        case 2:
            text += 'B';
            break;
    }
    return text;
}
console.log(describe(1));`)
}

func TestSwitchEmptyElseFallsThrough(test *testing.T) {
	test.Parallel()
	lowersAndAgreesWithNode(test, `function stop(): boolean { return false; }
function describe(value: number): string {
    let text = '';
    switch (value) {
        case 1:
            text += 'A';
            if (stop()) { break; } else {}
        case 2:
            text += 'B';
            break;
    }
    return text;
}
console.log(describe(1));`)
}

func TestSwitchEmptyCaseFallsThrough(test *testing.T) {
	test.Parallel()
	lowersAndAgreesWithNode(test, `function describe(value: number): string {
    let text = 'A';
    switch (value) {
        case 0:
        case 1: {}
        case 2:
            text += 'B';
            break;
    }
    return text;
}
console.log(describe(0));`)
}

func TestSwitchEmptyBlockFallsThrough(test *testing.T) {
	test.Parallel()
	lowersAndAgreesWithNode(test, `function describe(value: number): string {
    let text = '';
    switch (value) {
        case 1:
            text += 'A';
            {}
        case 2:
            text += 'B';
            break;
    }
    return text;
}
console.log(describe(1));`)
}

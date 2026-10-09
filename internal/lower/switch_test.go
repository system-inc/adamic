package lower

import "testing"

func TestSwitchConditionalBreakFallsThrough(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, `function stop(): boolean { return false; }
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

func TestSwitchEmptyElseFallsThrough(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, `function stop(): boolean { return false; }
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

func TestSwitchEmptyCaseFallsThrough(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, `function describe(value: number): string {
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

func TestSwitchEmptyBlockFallsThrough(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, `function describe(value: number): string {
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

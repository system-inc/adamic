package lower

import (
	"strings"
	"testing"
)

func TestHostErrorToStringIsNotYet(t *testing.T) {
	_, err := lowerSource(t, `import {openSync} from 'node:fs'; try { openSync('/missing', 'r'); } catch (error) { if (error instanceof Error) { console.log(error.toString()); } }`)
	if err == nil || !strings.Contains(err.Error(), "adamic/host-error-to-string") {
		t.Fatalf("want named host Error.toString boundary, got %v", err)
	}
}

func TestHostErrorStandardFormattingCompiles(t *testing.T) {
	_, err := lowerSource(t, `import {openSync} from 'node:fs'; try { openSync('/missing', 'r'); } catch (error) { if (error instanceof Error) { console.log(Error.prototype.toString.call(error)); } }`)
	if err != nil {
		t.Fatal(err)
	}
}

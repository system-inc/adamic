package lower

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func checkEepPresenceRefusal(test *testing.T, name, message string) {
	test.Helper()
	source, readError := os.ReadFile("testdata/eep_presence/" + name + ".a")
	if readError != nil {
		test.Fatal(readError)
	}
	_, lowerSourceError := lowerSource(test, string(source))
	var notYet *NotYet
	if !errors.As(lowerSourceError, &notYet) ||
		!strings.Contains(notYet.What, message) ||
		!strings.Contains(notYet.Where, "main.a:") {
		test.Fatalf("want located presence refusal with workaround, got %v", lowerSourceError)
	}
}

func TestEepPresence0(test *testing.T) {
	test.Parallel()
	checkEepPresenceRefusal(test, "eep1m5z", "union or optional tuple length")
}

func TestEepPresence1(test *testing.T) {
	test.Parallel()
	checkEepPresenceRefusal(test, "eep1m5z-93122fa_t_t5", "union or optional tuple length")
}

func TestEepPresence2(test *testing.T) {
	test.Parallel()
	checkEepPresenceRefusal(test, "eep1m5z-93122fa_t_v1", "union or optional tuple length")
}

func TestEepPresence3(test *testing.T) {
	test.Parallel()
	checkEepPresenceRefusal(test, "eep1m5z-93122fa_t_v2", "union or optional tuple length")
}

func TestEepPresence4(test *testing.T) {
	test.Parallel()
	checkEepPresenceRefusal(test, "eep1m5z-93122fa_t_v5", "union or optional tuple length")
}

func TestEepPresence5(test *testing.T) {
	test.Parallel()
	checkEepPresenceRefusal(test, "eep1m5z-classfeat_method_view", "method sharing its name with a getter")
}

func TestEepPresence6(test *testing.T) {
	test.Parallel()
	checkEepPresenceRefusal(
		test,
		"eep1m5z-classfeat_static_virtual",
		"static field read with an earlier initializer calling this",
	)
}

func TestEepPresenceSupported(test *testing.T) {
	test.Parallel()
	_, lowerSourceError := lowerSource(
		test,
		`const pair: [string, number] = ['a', 1]; console.log(String(pair.length)); const values: number[] = [1,2]; console.log(String(values?.length)); class Safe { static label = 'ok'; static early = this.read(); static read(): string { return this.label; } } console.log(Safe.early);`,
	)
	if lowerSourceError != nil {
		test.Fatal(lowerSourceError)
	}
}

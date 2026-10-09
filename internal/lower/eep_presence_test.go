package lower

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func checkEEPPresenceRefusal(t *testing.T, name, message string) {
	t.Helper()
	source, err := os.ReadFile("testdata/eep_presence/" + name + ".a")
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowerSource(t, string(source))
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(notYet.What, message) || !strings.Contains(notYet.Where, "main.a:") {
		t.Fatalf("want located presence refusal with workaround, got %v", err)
	}
}

func TestEEPPresence0(t *testing.T) {
	t.Parallel()
	checkEEPPresenceRefusal(t, "eep1m5z", "union or optional tuple length")
}

func TestEEPPresence1(t *testing.T) {
	t.Parallel()
	checkEEPPresenceRefusal(t, "eep1m5z-93122fa_t_t5", "union or optional tuple length")
}

func TestEEPPresence2(t *testing.T) {
	t.Parallel()
	checkEEPPresenceRefusal(t, "eep1m5z-93122fa_t_v1", "union or optional tuple length")
}

func TestEEPPresence3(t *testing.T) {
	t.Parallel()
	checkEEPPresenceRefusal(t, "eep1m5z-93122fa_t_v2", "union or optional tuple length")
}

func TestEEPPresence4(t *testing.T) {
	t.Parallel()
	checkEEPPresenceRefusal(t, "eep1m5z-93122fa_t_v5", "union or optional tuple length")
}

func TestEEPPresence5(t *testing.T) {
	t.Parallel()
	checkEEPPresenceRefusal(t, "eep1m5z-classfeat_method_view", "method sharing its name with a getter")
}

func TestEEPPresence6(t *testing.T) {
	t.Parallel()
	checkEEPPresenceRefusal(t, "eep1m5z-classfeat_static_virtual", "static field read with an earlier initializer calling this")
}

func TestEEPPresenceSupported(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `const pair: [string, number] = ['a', 1]; console.log(String(pair.length)); const values: number[] = [1,2]; console.log(String(values?.length)); class Safe { static label = 'ok'; static early = this.read(); static read(): string { return this.label; } } console.log(Safe.early);`)
	if err != nil {
		t.Fatal(err)
	}
}

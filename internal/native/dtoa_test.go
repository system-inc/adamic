package native

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// dtoaHarness formats one number per line: which method, whether there's an argument, the argument
// and the value as hexadecimal bits. It writes the string the method returns.
const dtoaHarness = `#include "adamic.h"

#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static double read(const char *text) {
	uint64_t bits = strtoull(text, NULL, 16);
	double value;
	memcpy(&value, &bits, sizeof value);
	return value;
}

int main(void) {
	char operation[16], digits[32], value[32];
	int has_digits;
	while (scanf("%15s %d %31s %31s", operation, &has_digits, digits, value) == 4) {
		adamic_string *text = strcmp(operation, "exponential") == 0
			? adamic_number_to_exponential(read(value), read(digits), has_digits != 0)
			: adamic_number_to_precision(read(value), read(digits), has_digits != 0);
		fwrite(text->bytes, 1, text->length, stdout);
		fputc('\n', stdout);
		adamic_release(text);
	}
	return 0;
}
`

const dtoaOracle = `const view = new DataView(new ArrayBuffer(8));
const read = (text) => { view.setBigUint64(0, BigInt('0x' + text)); return view.getFloat64(0); };
const out = [];
for (const line of require('node:fs').readFileSync(0, 'utf8').trim().split('\n')) {
	const [operation, hasDigits, digits, value] = line.split(' ');
	const number = read(value);
	const argument = hasDigits === '1' ? [read(digits)] : [];
	out.push(operation === 'exponential' ? number.toExponential(...argument) : number.toPrecision(...argument));
}
process.stdout.write(out.join('\n') + '\n');
`

// dtoaValues are the numbers both methods are tried on: the edges, the ties that round half up in
// JavaScript and half to even in printf, numbers people write, and any bits at all.
func dtoaValues(random *rand.Rand, count int) []float64 {
	values := []float64{
		0,
		math.Copysign(0, -1),
		math.NaN(),
		math.Inf(1),
		math.Inf(-1),
		math.SmallestNonzeroFloat64,
		math.Float64frombits(0x000FFFFFFFFFFFFF),
		math.Float64frombits(0x0010000000000000),
		math.MaxFloat64,
		1,
		2.5,
		0.5,
		1.5,
		0.125,
		0.375,
		1.25,
		12.5,
		125,
		1e21,
		1e-7,
		1e-6,
		123.456,
		0.1,
		0.2,
		0.3,
		1.005,
		1.45,
		8.345,
		9.5,
		99.5,
		999.5,
		9.95,
		0.95,
		0.000095,
		9.999999999999999e22,
		1e23,
		5e-324,
		2.225e-308,
		4503599627370496,
		9007199254740992,
		9007199254740993,
		1.7976931348623157e308,
		26.568583470577035,
		math.Pi,
		math.E,
		1e100,
		1e-100,
		123456789012345680000,
		0.000001234,
		1234.5678,
		5,
		50,
		55,
		0.05,
		0.55,
		0.555,
		1.0000000000000002,
		// Exact ties for the shortest digits, which round half to even there: 2^-25 is 2.9802322387695312e-8,
		// not ...313. Of every odd multiple below 8 of every power of two, only these four are ties, and
		// the random values found just one more (1.7860489244912068e15).
		math.Ldexp(1, -25),
		math.Ldexp(3, -24),
		math.Ldexp(5, -23),
		math.Ldexp(7, -23),
		1.7860489244912068e15,
	}
	for exponent := -30; exponent <= 30; exponent++ {
		values = append(
			values,
			math.Pow(10, float64(exponent)),
			5*math.Pow(10, float64(exponent)),
			9.5*math.Pow(10, float64(exponent)),
		)
	}
	for range count {
		values = append(values, math.Float64frombits(random.Uint64()))
		values = append(values, float64(random.IntN(2_000_000)-1_000_000)/math.Pow(10, float64(random.IntN(9))))
		// A tie at some place: an odd count of halves.
		values = append(values, float64(2*random.IntN(100_000)+1)/math.Pow(2, float64(1+random.IntN(12))))
		values = append(values, math.Ldexp(random.Float64()+0.5, random.IntN(2100)-1075))
	}
	all := []float64{}
	for _, value := range values {
		all = append(all, value, -value)
	}
	return all
}

// dtoaSweepValues are the values the sweep asks about, the same list on every call.
func dtoaSweepValues() []float64 {
	return dtoaValues(rand.New(rand.NewPCG(2011, 1993)), 400)
}

// The values the sweep is cut into, by their places in dtoaSweepValues (first to last, both
// included). Each is asked about with every argument, and the whole list is more than one test
// should carry alone, so each range is its own top-level test below, and
// TestToExponentialAndToPrecisionValueRangesTile holds the ranges to every value once and to their tests.
var dtoaValueRanges = map[string]testUnitRange{
	"0 to 922":     {0, 923},
	"923 to 1844":  {923, 1845},
	"1845 to 2767": {1845, 2768},
	"2768 to 3689": {2768, 3690},
}

func TestToExponentialAndToPrecisionMatchNodeValues0To922(test *testing.T) {
	test.Parallel()
	runDtoaValues(test, "0 to 922")
}

func TestToExponentialAndToPrecisionMatchNodeValues923To1844(test *testing.T) {
	test.Parallel()
	runDtoaValues(test, "923 to 1844")
}

func TestToExponentialAndToPrecisionMatchNodeValues1845To2767(test *testing.T) {
	test.Parallel()
	runDtoaValues(test, "1845 to 2767")
}

func TestToExponentialAndToPrecisionMatchNodeValues2768To3689(test *testing.T) {
	test.Parallel()
	runDtoaValues(test, "2768 to 3689")
}

// A range that leaves a value out or asks about one twice, a range without its test, or a test
// naming a range that isn't there fails here without building anything.
func TestToExponentialAndToPrecisionValueRangesTile(test *testing.T) {
	test.Parallel()
	checkRangesTile(test, "%d to %d", len(dtoaSweepValues()), dtoaValueRanges)
	checkNamedWrappers(
		test,
		"dtoa_test.go",
		"TestToExponentialAndToPrecisionMatchNodeValues",
		"runDtoaValues",
		dtoaValueRanges,
	)
}

// dtoaHarnessBinary is the harness every range of the sweep runs, built once for all of them in
// this process, in a directory of its own, since a test's own would go when that test ends.
var dtoaHarnessBinary = sync.OnceValues(func() (string, error) {
	directory, temporaryDirectoryError := os.MkdirTemp("", "adamic-dtoa-")
	if temporaryDirectoryError != nil {
		return "", temporaryDirectoryError
	}
	binary := filepath.Join(directory, "harness")
	return binary, Build(dtoaHarness, binary, Options{Sanitize: true})
})

func runDtoaValues(test *testing.T, name string) {
	test.Helper()
	piece, ok := dtoaValueRanges[name]
	if !ok {
		test.Fatalf("no value range %q", name)
	}
	values := dtoaSweepValues()[piece.first:piece.last]
	var input strings.Builder
	bits := func(value float64) string { return fmt.Sprintf("%016x", math.Float64bits(value)) }
	lines := 0
	ask := func(operation string, hasDigits bool, digits float64, value float64) {
		has := 0
		if hasDigits {
			has = 1
		}
		fmt.Fprintf(&input, "%s %d %s %s\n", operation, has, bits(digits), bits(value))
		lines++
	}
	for _, value := range values {
		ask("exponential", false, 0, value)
		ask("precision", false, 0, value)
		// Every argument there is: 0 to 100 digits after the point, 1 to 100 significant digits.
		for digits := 0; digits <= 100; digits++ {
			ask("exponential", true, float64(digits), value)
			if digits >= 1 {
				ask("precision", true, float64(digits), value)
			}
		}
		// Arguments that ToIntegerOrInfinity turns into ones in range.
		for _, digits := range []float64{2.7, -0.5, math.NaN(), 100.9, 1.5} {
			ask("exponential", true, digits, value)
			if digits >= 1 {
				ask("precision", true, digits, value)
			}
		}
		// NaN and the infinities answer before the argument is checked, so these don't throw.
		if math.IsNaN(value) || math.IsInf(value, 0) {
			for _, digits := range []float64{-1, 101, 1000, math.Inf(1), math.Inf(-1), 0} {
				ask("exponential", true, digits, value)
				ask("precision", true, digits, value)
			}
		}
	}

	binary, dtoaHarnessBinaryError := dtoaHarnessBinary()
	if dtoaHarnessBinaryError != nil {
		test.Fatal(dtoaHarnessBinaryError)
	}
	native := strings.Split(strings.TrimSuffix(runWithInput(test, input.String(), binary), "\n"), "\n")
	oracle := strings.Split(
		strings.TrimSuffix(runWithInput(test, input.String(), "node", "--eval", dtoaOracle), "\n"),
		"\n",
	)
	asked := strings.Split(strings.TrimSuffix(input.String(), "\n"), "\n")
	if len(native) != lines || len(oracle) != lines {
		test.Fatalf("got %d native and %d Node answers for %d questions", len(native), len(oracle), lines)
	}
	mismatches := map[string]int{}
	total := 0
	for index := range asked {
		if native[index] != oracle[index] {
			operation, _, _ := strings.Cut(asked[index], " ")
			mismatches[operation]++
			total++
			if total <= 20 {
				test.Errorf("%s: native %q, Node %q", asked[index], native[index], oracle[index])
			}
		}
	}
	if total > 0 {
		test.Errorf("%d of %d answers differ, by method: %v", total, lines, mismatches)
	}
	test.Logf("%d answers (%d values, every argument), %d mismatches", lines, len(values), total)
}

// Out of range, Node throws a RangeError, and so does Adamic: the call leaves the Error on the pending
// word, its name and message apart, and returns an empty string the caller lets go. Nothing catches
// it here, so it ends as an uncaught error does, exit 70 with stdout flushed. Each case is its own
// process, since that ends it.
func TestToExponentialAndToPrecisionOutOfRangeThrow(test *testing.T) {
	test.Parallel()
	const harness = `#include "adamic.h"

#include <math.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int main(int count, char **arguments) {
	(void)count;
	double digits = strtod(arguments[2], NULL);
	// Written as a program writes, so the uncaught error is what flushes it.
	static adamic_string before = ADAMIC_STRING("before");
	adamic_write_line(adamic_stdout, &before);
	adamic_string *text = strcmp(arguments[1], "exponential") == 0 ? adamic_number_to_exponential(1.5, digits, true) : adamic_number_to_precision(1.5, digits, true);
	if (adamic_thrown != NULL) {
		const adamic_string *name = adamic_thrown->slots[0].reference;
		const adamic_string *message = adamic_thrown->slots[1].reference;
		fprintf(stderr, "thrown %.*s|%.*s|%zu\n", (int)name->length, name->bytes, (int)message->length, message->bytes, text->length);
		adamic_release(text);
		adamic_uncaught();
	}
	adamic_write_line(adamic_stdout, text);
	adamic_release(text);
	return 0;
}
`
	binary := filepath.Join(test.TempDir(), "harness")
	if buildError := Build(harness, binary, Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	for _, probe := range []struct {
		operation string
		digits    string
		throws    bool
	}{
		{"exponential", "-1", true},
		{"exponential", "101", true},
		{"exponential", "inf", true},
		{"exponential", "-inf", true},
		{"exponential", "-0.9", false},
		{"exponential", "100.5", false},
		{"precision", "0", true},
		{"precision", "0.9", true},
		{"precision", "101", true},
		{"precision", "-5", true},
		{"precision", "inf", true},
		{"precision", "1", false},
		{"precision", "100", false},
	} {
		method := map[string]string{"exponential": "toExponential", "precision": "toPrecision"}[probe.operation]
		message := map[string]string{
			"exponential": "toExponential() argument must be between 0 and 100",
			"precision":   "toPrecision() argument must be between 1 and 100",
		}[probe.operation]
		digits := map[string]string{"inf": "Infinity", "-inf": "-Infinity"}[probe.digits]
		if digits == "" {
			digits = probe.digits
		}
		script := fmt.Sprintf("console.log('before'); console.log((1.5).%s(%s));", method, digits)
		native := outcome(test, binary, probe.operation, probe.digits)
		node := outcome(test, "node", "--eval", script)
		if native.stdout != node.stdout ||
			(native.exitCode == 70) != probe.throws ||
			(node.exitCode != 0) != probe.throws {
			test.Errorf(
				"(1.5).%s(%s): native exit %d %q, Node exit %d %q",
				method,
				digits,
				native.exitCode,
				native.stdout,
				node.exitCode,
				node.stdout,
			)
		}
		// The name and the message are the Error's two fields, never one text in the message.
		thrown := "thrown RangeError|" + message + "|0\n"
		if probe.throws && (!strings.Contains(native.stderr, thrown) ||
			!strings.Contains(native.stderr, "adamic: panic: RangeError: "+message) ||
			!strings.Contains(node.stderr, "RangeError: "+message)) {
			test.Errorf(
				"(1.5).%s(%s): native stderr %q, want %q; Node stderr %q",
				method,
				digits,
				native.stderr,
				thrown,
				node.stderr,
			)
		}
	}
}

// ran is how a process ended: what it wrote and its exit code.
type ran struct {
	stdout, stderr string
	exitCode       int
}

// outcome runs a program to the end, a nonzero exit included, which runWithInput takes as failure.
func outcome(test *testing.T, name string, arguments ...string) ran {
	test.Helper()
	timeoutContext, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	command := programCommand(timeoutContext, "", name, arguments...)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	runError := command.Run()
	var exit *exec.ExitError
	if runError != nil && !errors.As(runError, &exit) {
		test.Fatalf("%s: %v", name, runError)
	}
	return ran{stdout.String(), stderr.String(), command.ProcessState.ExitCode()}
}

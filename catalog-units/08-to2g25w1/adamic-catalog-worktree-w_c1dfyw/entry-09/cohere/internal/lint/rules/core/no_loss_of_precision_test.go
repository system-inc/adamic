package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// lossOfPrecisionFile is where the fixtures pretend to live.
const lossOfPrecisionFile = "/repository/source/Numbers.ts"

// The corpus is oxc's, copied rather than rewritten.
//
// One Tester block, 89 pass and 56 fail, and the snapshot records 56 diagnostics from 56 fail
// inputs. One finding per input holds throughout, which is unusual enough to be worth stating: no
// input here reports twice, so every fail case asserts exactly one message id.
func TestNoLossOfPrecisionFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"9007199254740993", "var x = 9007199254740993"},
		{"9007199254740.993e3", "var x = 9007199254740.993e3"},
		{"9.007199254740993e15", "var x = 9.007199254740993e15"},
		{"-9007199254740993", "var x = -9007199254740993"},
		{"900719.9254740994", "var x = 900719.9254740994"},
		{"-900719.9254740994", "var x = -900719.9254740994"},
		{"-9_00719_9254_740993", "const x = -9_00719_9254_740993"},
		{"-900_719.92_5474_0994", "const x = -900_719.92_5474_0994"},
		{"9.0_0719925_474099_3e15", "const x = 9.0_0719925_474099_3e15"},
		{"900_719.92_54740_994", "const x = 900_719.92_54740_994"},
		{"-9_00719_9254_740993 #2", "let x = -9_00719_9254_740993"},
		{"-900_719.92_5474_0994 #2", "let x = -900_719.92_5474_0994"},
		{"9.0_0719925_474099_3e15 #2", "let x = 9.0_0719925_474099_3e15"},
		{"900_719.92_54740_994 #2", "let x = 900_719.92_54740_994"},
		{"-9_00719_9254_740993 #3", "var x = -9_00719_9254_740993"},
		{"-900_719.92_5474_0994 #3", "var x = -900_719.92_5474_0994"},
		{".9007199254740993e16", "var x = .9007199254740993e16"},
		{"9.0_0719925_474099_3e15 #3", "var x = 9.0_0719925_474099_3e15"},
		{"90_0719925_4740.9_93e3", "var x = 90_0719925_4740.9_93e3"},
		{"900_719.92_54740_994 #3", "var x = 900_719.92_54740_994"},
		{"900719925474099_3", "var x = 900719925474099_3"},
		{"900719925474099.30e1", "var x = 900719925474099.30e1"},
		{"90071992547409930e-1", "var x = 90071992547409930e-1"},
		{"5123000000000000000000000000001", "var x = 5123000000000000000000000000001"},
		{"-5123000000000000000000000000001", "var x = -5123000000000000000000000000001"},
		{"1230000000000000000000000.0", "var x = 1230000000000000000000000.0"},
		{"1.0000000000000000000000123", "var x = 1.0000000000000000000000123"},
		{"1749800579826409539498001781694097092282535...", "var x = 17498005798264095394980017816940970922825355447145699491406164851279623993595007385788105416184430592"},
		{"2e999", "var x = 2e999"},
		{".1230000000000000000000000", "var x = .1230000000000000000000000"},
		{"0b10000000000000000000000000000000000000000...", "var x = 0b100000000000000000000000000000000000000000000000000001"},
		{"0B10000000000000000000000000000000000000000...", "var x = 0B100000000000000000000000000000000000000000000000000001"},
		{"0o400000000000000001", "var x = 0o400000000000000001"},
		{"0O400000000000000001", "var x = 0O400000000000000001"},
		{"0400000000000000001", "var x = 0400000000000000001"},
		{"0x20000000000001", "var x = 0x20000000000001"},
		{"0X20000000000001", "var x = 0X20000000000001"},
		{"5123_00000000000000000000000000_1", "var x = 5123_00000000000000000000000000_1"},
		{"-5_12300000000000000000000_0000001", "var x = -5_12300000000000000000000_0000001"},
		{"123_00000000000000000000_00.0_0", "var x = 123_00000000000000000000_00.0_0"},
		{"1.0_00000000000000000_0000123", "var x = 1.0_00000000000000000_0000123"},
		{"174_980057982_640953949800178169_4097092282...", "var x = 174_980057982_640953949800178169_409709228253554471456994_914061648512796239935950073857881054_1618443059_2"},
		{"2e9_99", "var x = 2e9_99"},
		{".1_23000000000000_00000_0000_0", "var x = .1_23000000000000_00000_0000_0"},
		{"0b1_000000000000000000000000000000000000000...", "var x = 0b1_0000000000000000000000000000000000000000000000000000_1"},
		{"0B10000000000_0000000000000000000000000000_...", "var x = 0B10000000000_0000000000000000000000000000_000000000000001"},
		{"0o4_00000000000000_001", "var x = 0o4_00000000000000_001"},
		{"0O4_0000000000000000_1", "var x = 0O4_0000000000000000_1"},
		{"0x2_0000000000001", "var x = 0x2_0000000000001"},
		{"0X200000_0000000_1", "var x = 0X200000_0000000_1"},
		{"9007199254740993 #2", "const x = 9007199254740993;"},
		{"9_007_199_254_740_993", "const x = 9_007_199_254_740_993;"},
		{"9_007_199_254_740.993e3", "const x = 9_007_199_254_740.993e3;"},
		{"0b100_000_000_000_000_000_000_000_000_000_0...", "const x = 0b100_000_000_000_000_000_000_000_000_000_000_000_000_000_000_000_000_001;"},
		{"1e18_446_744_073_709_551_615", "var x = 1e18_446_744_073_709_551_615"},
		{"96215808661.52751e-84", "var x = 96215808661.52751e-84"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoLossOfPrecision, lossOfPrecisionFile, testCase.sourceText),
				"noLossOfPrecision")
		})
	}
}

// The clean cases carry the whole discrimination and most of them exist because a plausible
// implementation reports them.
//
// `123.0000000000000000000000` and `123.000_000_000_000_000_000_000_0` are 25 and 22 digits of
// which only three are significant, so a rule counting digits rather than comparing round-trips
// reports both. `0195`, `019.5`, and `0008` look like octal and are not, because a legacy octal
// literal only exists when every digit is octal. `0x1FFF_FFFF_FFF_FFF` sits exactly at 53 bits and
// its neighbour `0x20000000000001` is the first hex value that does not fit. And `3e-308` is
// subnormal territory, where a formatter with a fixed fractional width runs out of digits before it
// reaches a significant one.
func TestNoLossOfPrecisionStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"12345", "var x = 12345"},
		{"123.456", "var x = 123.456"},
		{"-123.456", "var x = -123.456"},
		{"-123456", "var x = -123456"},
		{"123e34", "var x = 123e34"},
		{"123.0e34", "var x = 123.0e34"},
		{"123e-34", "var x = 123e-34"},
		{"-123e34", "var x = -123e34"},
		{"-123e-34", "var x = -123e-34"},
		{"12.3e34", "var x = 12.3e34"},
		{"12.3e-34", "var x = 12.3e-34"},
		{"-12.3e34", "var x = -12.3e34"},
		{"-12.3e-34", "var x = -12.3e-34"},
		{"12300000000000000000000000", "var x = 12300000000000000000000000"},
		{"-12300000000000000000000000", "var x = -12300000000000000000000000"},
		{"0.00000000000000000000000123", "var x = 0.00000000000000000000000123"},
		{"-0.00000000000000000000000123", "var x = -0.00000000000000000000000123"},
		{"9007199254740991", "var x = 9007199254740991"},
		{"0", "var x = 0"},
		{"0.0", "var x = 0.0"},
		{"0.00000000000000000000000000000000000000000...", "var x = 0.000000000000000000000000000000000000000000000000000000000000000000000000000000"},
		{"-0", "var x = -0"},
		{"123.0000000000000000000000", "var x = 123.0000000000000000000000"},
		{"9.00e2", "var x = 9.00e2"},
		{"9.000e3", "var x = 9.000e3"},
		{"9.0000000000e10", "var x = 9.0000000000e10"},
		{"9.00E2", "var x = 9.00E2"},
		{"9.000E3", "var x = 9.000E3"},
		{"9.100E3", "var x = 9.100E3"},
		{"9.0000000000E10", "var x = 9.0000000000E10"},
		{"019.5", "var x = 019.5"},
		{"0195", "var x = 0195"},
		{"00195", "var x = 00195"},
		{"0008", "var x = 0008"},
		{"0e5", "var x = 0e5"},
		{".42", "var x = .42"},
		{"42.", "var x = 42."},
		{"12_34_56", "var x = 12_34_56"},
		{"12_3.4_56", "var x = 12_3.4_56"},
		{"-12_3.4_56", "var x = -12_3.4_56"},
		{"-12_34_56", "var x = -12_34_56"},
		{"12_3e3_4", "var x = 12_3e3_4"},
		{"123.0e3_4", "var x = 123.0e3_4"},
		{"12_3e-3_4", "var x = 12_3e-3_4"},
		{"12_3.0e-3_4", "var x = 12_3.0e-3_4"},
		{"-1_23e-3_4", "var x = -1_23e-3_4"},
		{"-1_23.8e-3_4", "var x = -1_23.8e-3_4"},
		{"1_230000000_00000000_00000_000", "var x = 1_230000000_00000000_00000_000"},
		{"-1_230000000_00000000_00000_000", "var x = -1_230000000_00000000_00000_000"},
		{"0.0_00_000000000_000000000_00123", "var x = 0.0_00_000000000_000000000_00123"},
		{"-0.0_00_000000000_000000000_00123", "var x = -0.0_00_000000000_000000000_00123"},
		{"0e5_3", "var x = 0e5_3"},
		{"0b11111111111111111111111111111111111111111...", "var x = 0b11111111111111111111111111111111111111111111111111111"},
		{"0b111_111_111_111_1111_11111_111_11111_1111...", "var x = 0b111_111_111_111_1111_11111_111_11111_1111111111_11111111_111_111"},
		{"0B11111111111111111111111111111111111111111...", "var x = 0B11111111111111111111111111111111111111111111111111111"},
		{"0B111_111_111_111_1111_11111_111_11111_1111...", "var x = 0B111_111_111_111_1111_11111_111_11111_1111111111_11111111_111_111"},
		{"0o377777777777777777", "var x = 0o377777777777777777"},
		{"0o3_77_777_777_777_777_777", "var x = 0o3_77_777_777_777_777_777"},
		{"0O377777777777777777", "var x = 0O377777777777777777"},
		{"0x1FFFFFFFFFFFFF", "var x = 0x1FFFFFFFFFFFFF"},
		{"0X1FFFFFFFFFFFFF", "var x = 0X1FFFFFFFFFFFFF"},
		{"true", "var x = true"},
		{"'abc'", "var x = 'abc'"},
		{"''", "var x = ''"},
		{"null", "var x = null"},
		{"undefined", "var x = undefined"},
		{"{}", "var x = {}"},
		{"['a', 'b']", "var x = ['a', 'b']"},
		{"new Date()", "var x = new Date()"},
		{"'9007199254740993'", "var x = '9007199254740993'"},
		{"0x1FFF_FFFF_FFF_FFF", "var x = 0x1FFF_FFFF_FFF_FFF"},
		{"0X1_FFF_FFFF_FFF_FFF", "var x = 0X1_FFF_FFFF_FFF_FFF"},
		{"12345 #2", "const x = 12345;"},
		{"123.456 #2", "const x = 123.456;"},
		{"-123.456 #2", "const x = -123.456;"},
		{"123_456", "const x = 123_456;"},
		{"123_00_000_000_000_000_000_000_000", "const x = 123_00_000_000_000_000_000_000_000;"},
		{"123.000_000_000_000_000_000_000_0", "const x = 123.000_000_000_000_000_000_000_0;"},
		{"Infinity", "var a = Infinity"},
		{"480.00", "var a = 480.00"},
		{"-30.00", "var a = -30.00"},
		{"Infinity #2", "let a = Infinity"},
		{"480.00 #2", "let a = 480.00"},
		{"-30.00 #2", "let a = -30.00"},
		{"Infinity #3", "const a = Infinity"},
		{"480.00 #3", "const a = 480.00"},
		{"-30.00 #3", "const a = -30.00"},
		{"3e-308", "const x = 3e-308"},
		{"(1000000000000000128).toFixed(0)", "(1000000000000000128).toFixed(0)"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoLossOfPrecision, lossOfPrecisionFile, testCase.sourceText))
		})
	}
}

// Cases upstream does not cover, from reading our own code rather than oxc's.
//
// Upstream reads a `NumericLiteral` node whose `raw` is the source text. Our parser hands back a
// *normalized* `Text` instead, where `1e1` arrives as `10` and every non-decimal base arrives as
// its decimal value, so this rule reads the raw span out of the source file. These cases pin that
// decision: each one is a literal whose normalized text and source text differ, and a rule reading
// `.Text` reports the wrong answer on all of them.
func TestNoLossOfPrecisionReadsRawSourceText(t *testing.T) {
	t.Parallel()

	fires := []struct {
		name       string
		sourceText string
	}{
		// `.Text` for this is "9007199254740992", which round-trips exactly and would pass.
		{"a literal the parser has already rounded", "var x = 9007199254740993"},
		// `.Text` is the decimal "9007199254740992", which carries no hex digits to compare.
		{"hex whose normalized text is decimal", "var x = 0x20000000000001"},
	}
	for _, testCase := range fires {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoLossOfPrecision, lossOfPrecisionFile, testCase.sourceText),
				"noLossOfPrecision")
		})
	}

	silent := []struct {
		name       string
		sourceText string
	}{
		// A BigInt literal is a different node kind and is exempt: BigInt is exact at any width, so
		// there is no precision to lose. Upstream matches only `NumericLiteral` and reaches the
		// same exemption the same way. Without the exemption this is the clearest false positive
		// the rule could produce, because the digits are the canonical unsafe integer.
		{"a BigInt past the safe integer range", "var x = 9007199254740993n"},
		{"a BigInt in hex past 53 bits", "var x = 0x20000000000001n"},
		// The literal is inside a call rather than a declaration, which is where upstream's own
		// `(1000000000000000128).toFixed(0)` case lives. Position in the tree must not matter.
		{"a safe literal in a call argument", "callee(123.456);"},

		// These three exist because a mutation forcing the non-decimal round trip to report
		// survived the whole corpus: every non-decimal literal upstream ships is decided by the
		// bit-width check alone, so nothing measured whether the round trip can clear a literal.
		//
		// Each needs more than 53 bits and is still exact, because its low bits are zero. The
		// bit-width check counts positions rather than value and rejects all three, so the round
		// trip is the only thing that can pass them, and a rule whose round trip always reports
		// gets all three wrong while staying green on oxc's corpus.
		{"hex over 53 bits whose low bits are zero", "var x = 0x40000000000004"},
		{"octal over 53 bits whose low bits are zero", "var x = 0o1000000000000000000"},
		{"binary over 53 bits whose low bits are zero", "var x = 0b1100000000000000000000000000000000000000000000000000000"},
	}
	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoLossOfPrecision, lossOfPrecisionFile, testCase.sourceText))
		})
	}
}

package native

import (
	"path/filepath"
	"strings"
	"testing"
)

// Both programs generate the same reproducible 100,000 strings independently. Most are short;
// every thousandth has a long scrambled combining run. No huge integration probe is a fixture.
func TestNormalizeRandomMatchesNode(t *testing.T) {
	t.Parallel()
	alphabet := strings.Join(append(append([]string{}, normalizeAlphabet...), "0xfdfa", "0x315", "0x34f", "0x1f600", "0x9cb", "0x1ea0"), ", ")
	harness := strings.Split(normalizeHarness, "static const uint32_t alphabet[]")[0]
	first := strings.Index(harness, "static adamic_string *text(")
	last := strings.Index(harness[first:], "static adamic_string forms") + first
	harness = harness[:first] + `static adamic_string *text(const uint32_t *points, size_t count) {
 adamic_string *string = adamic_string_allocate(count * 4);
 size_t length = 0;
 for (size_t index = 0; index < count; index++) {
  length += encode(points[index], (char *)string->bytes + length);
 }
 string->length = adamic_string_join_halves((char *)string->bytes, 0, length);
 return string;
}

` + harness[last:] + `
static const uint32_t alphabet[] = {ALPHABET};
static const uint32_t marks[] = {0x300, 0x301, 0x323, 0x338, 0x345, 0x315};
static uint32_t seed = 0x243f6a88;
static uint32_t random_point(void) {
 seed ^= seed << 13;
 seed ^= seed >> 17;
 seed ^= seed << 5;
 return seed;
}
int main(void) {
 static char buffer[1 << 20];
 setvbuf(stdout, buffer, _IOFBF, sizeof buffer);
 for (size_t sample = 0; sample < 100000; sample++) {
  bool long_run = sample % 1000 == 0;
  size_t length = long_run ? 511 + random_point() % 3 : 1 + random_point() % 80;
  uint32_t points[514];
  for (size_t index = 0; index < length; index++) {
   uint32_t choice = random_point();
   points[index] = long_run ? marks[choice % 6] : alphabet[choice % (sizeof alphabet / sizeof alphabet[0])];
  }
  // Multiple runs, leading non-starters, and trailing marks, with Hangul, compatibility
  // decompositions and surrogate halves in the ordinary strings.
  if (long_run) { points[length / 2] = 'A'; }
  if (sample % 17 == 0) {
   size_t base = 1 + random_point() % (length < 8 ? length : 8);
   size_t repeated = base * (8 + random_point() % 24);
   for (size_t index = base; index < repeated; index++) { points[index] = points[index % base]; }
   length = repeated;
   for (size_t index = 0; index < 4; index++) { points[length++] = marks[random_point() % 6]; }
  }
  adamic_string *input = text(points, length);
  put_hex(input);
  adamic_release(input);
  answer(points, length);
  putchar('\n');
 }
 return 0;
}
`
	oracle := strings.Split(normalizeOracle, "const alphabet =")[0] + `
const alphabet = [ALPHABET];
const marks = [0x300, 0x301, 0x323, 0x338, 0x345, 0x315];
let seed = 0x243f6a88;
function randomPoint() {
 seed ^= seed << 13;
 seed ^= seed >>> 17;
 seed ^= seed << 5;
 return seed >>> 0;
}
let chunk = '';
for (let sample = 0; sample < 100000; sample++) {
 const longRun = sample % 1000 === 0;
 const length = longRun ? 511 + randomPoint() % 3 : 1 + randomPoint() % 80;
 const points = [];
 for (let index = 0; index < length; index++) {
  const choice = randomPoint();
  points.push(longRun ? marks[choice % 6] : alphabet[choice % alphabet.length]);
 }
 if (longRun) { points[Math.floor(length / 2)] = 0x41; }
 if (sample % 17 === 0) {
  const base = 1 + randomPoint() % Math.min(length, 8);
  const repeated = base * (8 + randomPoint() % 24);
  points.length = base;
  for (let index = base; index < repeated; index++) { points.push(points[index % base]); }
  for (let index = 0; index < 4; index++) { points.push(marks[randomPoint() % 6]); }
 }
 chunk += wtf8(points.map(character).join('')) + answer(points) + '\n';
 if (chunk.length > 1 << 20) { process.stdout.write(chunk); chunk = ''; }
}
process.stdout.write(chunk);
`
	binary := filepath.Join(t.TempDir(), "harness")
	if err := Build(strings.Replace(harness, "ALPHABET", alphabet, 1), binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	compareStreams(t, 100000, []string{binary}, []string{"node", "--eval", strings.Replace(oracle, "ALPHABET", alphabet, 1)})
}

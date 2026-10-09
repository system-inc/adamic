package markdownblocks

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"
)

func tokenizerEventLiveEnumeration(t *testing.T) (string, [][]uint16, []string, int) {
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	corpus, files := blockCorpus(t, root, "whitespace")
	// Upstream selection verifies the two fixed checkout pins against Git.
	// The live repository portion must independently remain non-empty.
	repositoryFiles, cohereFiles, checkerFiles := 0, 0, 0
	for _, item := range corpus {
		switch {
		case strings.HasPrefix(item.Name, "cohere/TypeScript/"):
			checkerFiles++
		case strings.HasPrefix(item.Name, "cohere/"):
			cohereFiles++
		case !strings.HasPrefix(item.Name, "generated/"):
			repositoryFiles++
		}
	}
	if repositoryFiles == 0 {
		t.Fatal("repository Markdown corpus is empty")
	}
	// Only pinned upstream counts are fixed: cohere 7945d102 and
	// TypeScript d92d9bfe, using auditCorpus's explicitly selected roots.
	if census() && (cohereFiles != 786 || checkerFiles != 67) {
		t.Fatalf("pinned upstream file counts: cohere %d/786, TypeScript %d/67", cohereFiles, checkerFiles)
	}
	inputs := [][]uint16{{}}
	names := []string{"empty"}
	for _, item := range corpus {
		inputs = append(inputs, utf16.Encode([]rune(item.Text)))
		names = append(names, item.Name)
	}
	for unit := 0; unit <= 65535; unit++ {
		inputs = append(inputs, []uint16{uint16(unit)})
		names = append(names, fmt.Sprintf("unit/%d", unit))
	}
	alphabet := []uint16{0, 9, 10, 13, 65279, 65}
	for length := 1; length <= 5; length++ {
		count := 1
		for i := 0; i < length; i++ {
			count *= len(alphabet)
		}
		for n := 0; n < count; n++ {
			units := make([]uint16, length)
			number := n
			for i := range units {
				units[i] = alphabet[number%len(alphabet)]
				number /= len(alphabet)
			}
			inputs = append(inputs, units)
			names = append(names, fmt.Sprintf("sequence/%d/%d", length, n))
		}
	}
	for _, prefix := range []string{"", "a", "ab", "abc", "abcd", "abcde", "中", "😀", "a😀"} {
		for _, tail := range []string{"\t", "\t\t", "\r\n", "\rX\n", "\rX", "\x00\t", "\ufeff\t"} {
			inputs = append(inputs, utf16.Encode([]rune(prefix+tail)))
			names = append(names, fmt.Sprintf("columns/%q/%q", prefix, tail))
		}
	}
	return root, inputs, names, files
}

func TestTokenizerEventsUnion(t *testing.T) {
	_, inputs, names, _ := tokenizerEventLiveEnumeration(t)
	shards := tokenizerEventShards(tokenizerEventKeys(names))
	if len(shards) != testTokenizerEventsShards {
		t.Fatal("static shard count mismatch")
	}
	validateTokenizerEventUnion(t, len(inputs), shards)
}

func TestTokenizerEvents_000(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 0) }
func TestTokenizerEvents_001(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 1) }
func TestTokenizerEvents_002(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 2) }
func TestTokenizerEvents_003(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 3) }
func TestTokenizerEvents_004(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 4) }
func TestTokenizerEvents_005(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 5) }
func TestTokenizerEvents_006(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 6) }
func TestTokenizerEvents_007(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 7) }
func TestTokenizerEvents_008(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 8) }
func TestTokenizerEvents_009(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 9) }
func TestTokenizerEvents_010(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 10) }
func TestTokenizerEvents_011(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 11) }
func TestTokenizerEvents_012(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 12) }
func TestTokenizerEvents_013(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 13) }
func TestTokenizerEvents_014(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 14) }
func TestTokenizerEvents_015(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 15) }
func TestTokenizerEvents_016(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 16) }
func TestTokenizerEvents_017(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 17) }
func TestTokenizerEvents_018(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 18) }
func TestTokenizerEvents_019(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 19) }
func TestTokenizerEvents_020(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 20) }
func TestTokenizerEvents_021(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 21) }
func TestTokenizerEvents_022(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 22) }
func TestTokenizerEvents_023(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 23) }
func TestTokenizerEvents_024(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 24) }
func TestTokenizerEvents_025(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 25) }
func TestTokenizerEvents_026(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 26) }
func TestTokenizerEvents_027(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 27) }
func TestTokenizerEvents_028(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 28) }
func TestTokenizerEvents_029(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 29) }
func TestTokenizerEvents_030(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 30) }
func TestTokenizerEvents_031(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 31) }
func TestTokenizerEvents_032(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 32) }
func TestTokenizerEvents_033(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 33) }
func TestTokenizerEvents_034(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 34) }
func TestTokenizerEvents_035(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 35) }
func TestTokenizerEvents_036(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 36) }
func TestTokenizerEvents_037(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 37) }
func TestTokenizerEvents_038(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 38) }
func TestTokenizerEvents_039(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 39) }
func TestTokenizerEvents_040(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 40) }
func TestTokenizerEvents_041(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 41) }
func TestTokenizerEvents_042(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 42) }
func TestTokenizerEvents_043(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 43) }
func TestTokenizerEvents_044(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 44) }
func TestTokenizerEvents_045(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 45) }
func TestTokenizerEvents_046(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 46) }
func TestTokenizerEvents_047(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 47) }
func TestTokenizerEvents_048(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 48) }
func TestTokenizerEvents_049(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 49) }
func TestTokenizerEvents_050(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 50) }
func TestTokenizerEvents_051(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 51) }
func TestTokenizerEvents_052(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 52) }
func TestTokenizerEvents_053(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 53) }
func TestTokenizerEvents_054(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 54) }
func TestTokenizerEvents_055(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 55) }
func TestTokenizerEvents_056(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 56) }
func TestTokenizerEvents_057(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 57) }
func TestTokenizerEvents_058(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 58) }
func TestTokenizerEvents_059(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 59) }
func TestTokenizerEvents_060(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 60) }
func TestTokenizerEvents_061(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 61) }
func TestTokenizerEvents_062(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 62) }
func TestTokenizerEvents_063(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 63) }
func TestTokenizerEvents_064(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 64) }
func TestTokenizerEvents_065(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 65) }
func TestTokenizerEvents_066(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 66) }
func TestTokenizerEvents_067(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 67) }
func TestTokenizerEvents_068(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 68) }
func TestTokenizerEvents_069(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 69) }
func TestTokenizerEvents_070(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 70) }
func TestTokenizerEvents_071(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 71) }
func TestTokenizerEvents_072(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 72) }
func TestTokenizerEvents_073(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 73) }
func TestTokenizerEvents_074(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 74) }
func TestTokenizerEvents_075(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 75) }
func TestTokenizerEvents_076(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 76) }
func TestTokenizerEvents_077(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 77) }
func TestTokenizerEvents_078(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 78) }
func TestTokenizerEvents_079(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 79) }
func TestTokenizerEvents_080(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 80) }
func TestTokenizerEvents_081(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 81) }
func TestTokenizerEvents_082(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 82) }
func TestTokenizerEvents_083(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 83) }
func TestTokenizerEvents_084(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 84) }
func TestTokenizerEvents_085(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 85) }
func TestTokenizerEvents_086(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 86) }
func TestTokenizerEvents_087(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 87) }
func TestTokenizerEvents_088(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 88) }
func TestTokenizerEvents_089(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 89) }
func TestTokenizerEvents_090(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 90) }
func TestTokenizerEvents_091(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 91) }
func TestTokenizerEvents_092(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 92) }
func TestTokenizerEvents_093(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 93) }
func TestTokenizerEvents_094(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 94) }
func TestTokenizerEvents_095(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 95) }
func TestTokenizerEvents_096(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 96) }
func TestTokenizerEvents_097(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 97) }
func TestTokenizerEvents_098(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 98) }
func TestTokenizerEvents_099(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 99) }
func TestTokenizerEvents_100(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 100) }
func TestTokenizerEvents_101(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 101) }
func TestTokenizerEvents_102(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 102) }
func TestTokenizerEvents_103(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 103) }
func TestTokenizerEvents_104(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 104) }
func TestTokenizerEvents_105(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 105) }
func TestTokenizerEvents_106(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 106) }
func TestTokenizerEvents_107(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 107) }
func TestTokenizerEvents_108(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 108) }
func TestTokenizerEvents_109(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 109) }
func TestTokenizerEvents_110(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 110) }
func TestTokenizerEvents_111(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 111) }
func TestTokenizerEvents_112(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 112) }
func TestTokenizerEvents_113(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 113) }
func TestTokenizerEvents_114(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 114) }
func TestTokenizerEvents_115(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 115) }
func TestTokenizerEvents_116(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 116) }
func TestTokenizerEvents_117(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 117) }
func TestTokenizerEvents_118(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 118) }
func TestTokenizerEvents_119(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 119) }
func TestTokenizerEvents_120(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 120) }
func TestTokenizerEvents_121(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 121) }
func TestTokenizerEvents_122(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 122) }
func TestTokenizerEvents_123(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 123) }
func TestTokenizerEvents_124(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 124) }
func TestTokenizerEvents_125(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 125) }
func TestTokenizerEvents_126(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 126) }
func TestTokenizerEvents_127(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 127) }
func TestTokenizerEvents_128(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 128) }
func TestTokenizerEvents_129(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 129) }
func TestTokenizerEvents_130(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 130) }
func TestTokenizerEvents_131(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 131) }
func TestTokenizerEvents_132(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 132) }
func TestTokenizerEvents_133(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 133) }
func TestTokenizerEvents_134(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 134) }
func TestTokenizerEvents_135(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 135) }
func TestTokenizerEvents_136(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 136) }
func TestTokenizerEvents_137(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 137) }
func TestTokenizerEvents_138(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 138) }
func TestTokenizerEvents_139(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 139) }
func TestTokenizerEvents_140(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 140) }
func TestTokenizerEvents_141(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 141) }
func TestTokenizerEvents_142(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 142) }
func TestTokenizerEvents_143(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 143) }
func TestTokenizerEvents_144(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 144) }
func TestTokenizerEvents_145(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 145) }
func TestTokenizerEvents_146(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 146) }
func TestTokenizerEvents_147(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 147) }
func TestTokenizerEvents_148(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 148) }
func TestTokenizerEvents_149(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 149) }
func TestTokenizerEvents_150(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 150) }
func TestTokenizerEvents_151(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 151) }
func TestTokenizerEvents_152(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 152) }
func TestTokenizerEvents_153(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 153) }
func TestTokenizerEvents_154(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 154) }
func TestTokenizerEvents_155(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 155) }
func TestTokenizerEvents_156(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 156) }
func TestTokenizerEvents_157(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 157) }
func TestTokenizerEvents_158(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 158) }
func TestTokenizerEvents_159(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 159) }
func TestTokenizerEvents_160(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 160) }
func TestTokenizerEvents_161(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 161) }
func TestTokenizerEvents_162(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 162) }
func TestTokenizerEvents_163(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 163) }
func TestTokenizerEvents_164(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 164) }
func TestTokenizerEvents_165(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 165) }
func TestTokenizerEvents_166(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 166) }
func TestTokenizerEvents_167(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 167) }
func TestTokenizerEvents_168(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 168) }
func TestTokenizerEvents_169(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 169) }
func TestTokenizerEvents_170(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 170) }
func TestTokenizerEvents_171(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 171) }
func TestTokenizerEvents_172(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 172) }
func TestTokenizerEvents_173(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 173) }
func TestTokenizerEvents_174(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 174) }
func TestTokenizerEvents_175(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 175) }
func TestTokenizerEvents_176(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 176) }
func TestTokenizerEvents_177(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 177) }
func TestTokenizerEvents_178(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 178) }
func TestTokenizerEvents_179(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 179) }
func TestTokenizerEvents_180(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 180) }
func TestTokenizerEvents_181(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 181) }
func TestTokenizerEvents_182(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 182) }
func TestTokenizerEvents_183(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 183) }
func TestTokenizerEvents_184(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 184) }
func TestTokenizerEvents_185(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 185) }
func TestTokenizerEvents_186(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 186) }
func TestTokenizerEvents_187(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 187) }
func TestTokenizerEvents_188(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 188) }
func TestTokenizerEvents_189(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 189) }
func TestTokenizerEvents_190(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 190) }
func TestTokenizerEvents_191(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 191) }
func TestTokenizerEvents_192(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 192) }
func TestTokenizerEvents_193(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 193) }
func TestTokenizerEvents_194(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 194) }
func TestTokenizerEvents_195(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 195) }
func TestTokenizerEvents_196(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 196) }
func TestTokenizerEvents_197(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 197) }
func TestTokenizerEvents_198(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 198) }
func TestTokenizerEvents_199(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 199) }
func TestTokenizerEvents_200(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 200) }
func TestTokenizerEvents_201(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 201) }
func TestTokenizerEvents_202(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 202) }
func TestTokenizerEvents_203(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 203) }
func TestTokenizerEvents_204(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 204) }
func TestTokenizerEvents_205(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 205) }
func TestTokenizerEvents_206(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 206) }
func TestTokenizerEvents_207(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 207) }
func TestTokenizerEvents_208(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 208) }
func TestTokenizerEvents_209(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 209) }
func TestTokenizerEvents_210(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 210) }
func TestTokenizerEvents_211(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 211) }
func TestTokenizerEvents_212(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 212) }
func TestTokenizerEvents_213(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 213) }
func TestTokenizerEvents_214(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 214) }
func TestTokenizerEvents_215(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 215) }
func TestTokenizerEvents_216(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 216) }
func TestTokenizerEvents_217(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 217) }
func TestTokenizerEvents_218(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 218) }
func TestTokenizerEvents_219(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 219) }
func TestTokenizerEvents_220(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 220) }
func TestTokenizerEvents_221(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 221) }
func TestTokenizerEvents_222(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 222) }
func TestTokenizerEvents_223(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 223) }
func TestTokenizerEvents_224(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 224) }
func TestTokenizerEvents_225(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 225) }
func TestTokenizerEvents_226(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 226) }
func TestTokenizerEvents_227(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 227) }
func TestTokenizerEvents_228(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 228) }
func TestTokenizerEvents_229(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 229) }
func TestTokenizerEvents_230(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 230) }
func TestTokenizerEvents_231(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 231) }
func TestTokenizerEvents_232(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 232) }
func TestTokenizerEvents_233(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 233) }
func TestTokenizerEvents_234(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 234) }
func TestTokenizerEvents_235(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 235) }
func TestTokenizerEvents_236(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 236) }
func TestTokenizerEvents_237(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 237) }
func TestTokenizerEvents_238(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 238) }
func TestTokenizerEvents_239(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 239) }
func TestTokenizerEvents_240(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 240) }
func TestTokenizerEvents_241(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 241) }
func TestTokenizerEvents_242(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 242) }
func TestTokenizerEvents_243(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 243) }
func TestTokenizerEvents_244(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 244) }
func TestTokenizerEvents_245(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 245) }
func TestTokenizerEvents_246(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 246) }
func TestTokenizerEvents_247(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 247) }
func TestTokenizerEvents_248(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 248) }
func TestTokenizerEvents_249(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 249) }
func TestTokenizerEvents_250(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 250) }
func TestTokenizerEvents_251(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 251) }
func TestTokenizerEvents_252(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 252) }
func TestTokenizerEvents_253(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 253) }
func TestTokenizerEvents_254(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 254) }
func TestTokenizerEvents_255(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 255) }
func TestTokenizerEvents_256(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 256) }
func TestTokenizerEvents_257(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 257) }
func TestTokenizerEvents_258(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 258) }
func TestTokenizerEvents_259(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 259) }
func TestTokenizerEvents_260(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 260) }
func TestTokenizerEvents_261(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 261) }
func TestTokenizerEvents_262(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 262) }
func TestTokenizerEvents_263(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 263) }
func TestTokenizerEvents_264(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 264) }
func TestTokenizerEvents_265(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 265) }
func TestTokenizerEvents_266(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 266) }
func TestTokenizerEvents_267(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 267) }
func TestTokenizerEvents_268(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 268) }
func TestTokenizerEvents_269(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 269) }
func TestTokenizerEvents_270(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 270) }
func TestTokenizerEvents_271(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 271) }
func TestTokenizerEvents_272(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 272) }
func TestTokenizerEvents_273(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 273) }
func TestTokenizerEvents_274(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 274) }
func TestTokenizerEvents_275(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 275) }
func TestTokenizerEvents_276(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 276) }
func TestTokenizerEvents_277(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 277) }
func TestTokenizerEvents_278(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 278) }
func TestTokenizerEvents_279(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 279) }
func TestTokenizerEvents_280(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 280) }
func TestTokenizerEvents_281(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 281) }
func TestTokenizerEvents_282(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 282) }
func TestTokenizerEvents_283(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 283) }
func TestTokenizerEvents_284(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 284) }
func TestTokenizerEvents_285(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 285) }
func TestTokenizerEvents_286(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 286) }
func TestTokenizerEvents_287(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 287) }
func TestTokenizerEvents_288(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 288) }
func TestTokenizerEvents_289(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 289) }
func TestTokenizerEvents_290(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 290) }
func TestTokenizerEvents_291(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 291) }
func TestTokenizerEvents_292(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 292) }
func TestTokenizerEvents_293(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 293) }
func TestTokenizerEvents_294(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 294) }
func TestTokenizerEvents_295(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 295) }
func TestTokenizerEvents_296(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 296) }
func TestTokenizerEvents_297(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 297) }
func TestTokenizerEvents_298(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 298) }
func TestTokenizerEvents_299(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 299) }
func TestTokenizerEvents_300(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 300) }
func TestTokenizerEvents_301(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 301) }
func TestTokenizerEvents_302(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 302) }
func TestTokenizerEvents_303(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 303) }
func TestTokenizerEvents_304(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 304) }
func TestTokenizerEvents_305(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 305) }
func TestTokenizerEvents_306(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 306) }
func TestTokenizerEvents_307(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 307) }
func TestTokenizerEvents_308(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 308) }
func TestTokenizerEvents_309(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 309) }
func TestTokenizerEvents_310(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 310) }
func TestTokenizerEvents_311(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 311) }
func TestTokenizerEvents_312(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 312) }
func TestTokenizerEvents_313(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 313) }
func TestTokenizerEvents_314(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 314) }
func TestTokenizerEvents_315(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 315) }
func TestTokenizerEvents_316(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 316) }
func TestTokenizerEvents_317(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 317) }
func TestTokenizerEvents_318(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 318) }
func TestTokenizerEvents_319(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 319) }
func TestTokenizerEvents_320(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 320) }
func TestTokenizerEvents_321(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 321) }
func TestTokenizerEvents_322(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 322) }
func TestTokenizerEvents_323(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 323) }
func TestTokenizerEvents_324(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 324) }
func TestTokenizerEvents_325(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 325) }
func TestTokenizerEvents_326(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 326) }
func TestTokenizerEvents_327(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 327) }
func TestTokenizerEvents_328(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 328) }
func TestTokenizerEvents_329(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 329) }
func TestTokenizerEvents_330(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 330) }
func TestTokenizerEvents_331(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 331) }
func TestTokenizerEvents_332(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 332) }
func TestTokenizerEvents_333(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 333) }
func TestTokenizerEvents_334(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 334) }
func TestTokenizerEvents_335(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 335) }
func TestTokenizerEvents_336(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 336) }
func TestTokenizerEvents_337(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 337) }
func TestTokenizerEvents_338(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 338) }
func TestTokenizerEvents_339(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 339) }
func TestTokenizerEvents_340(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 340) }
func TestTokenizerEvents_341(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 341) }
func TestTokenizerEvents_342(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 342) }
func TestTokenizerEvents_343(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 343) }
func TestTokenizerEvents_344(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 344) }
func TestTokenizerEvents_345(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 345) }
func TestTokenizerEvents_346(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 346) }
func TestTokenizerEvents_347(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 347) }
func TestTokenizerEvents_348(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 348) }
func TestTokenizerEvents_349(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 349) }
func TestTokenizerEvents_350(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 350) }
func TestTokenizerEvents_351(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 351) }
func TestTokenizerEvents_352(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 352) }
func TestTokenizerEvents_353(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 353) }
func TestTokenizerEvents_354(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 354) }
func TestTokenizerEvents_355(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 355) }
func TestTokenizerEvents_356(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 356) }
func TestTokenizerEvents_357(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 357) }
func TestTokenizerEvents_358(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 358) }
func TestTokenizerEvents_359(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 359) }
func TestTokenizerEvents_360(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 360) }
func TestTokenizerEvents_361(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 361) }
func TestTokenizerEvents_362(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 362) }
func TestTokenizerEvents_363(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 363) }
func TestTokenizerEvents_364(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 364) }
func TestTokenizerEvents_365(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 365) }
func TestTokenizerEvents_366(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 366) }
func TestTokenizerEvents_367(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 367) }
func TestTokenizerEvents_368(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 368) }
func TestTokenizerEvents_369(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 369) }
func TestTokenizerEvents_370(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 370) }
func TestTokenizerEvents_371(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 371) }
func TestTokenizerEvents_372(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 372) }
func TestTokenizerEvents_373(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 373) }
func TestTokenizerEvents_374(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 374) }
func TestTokenizerEvents_375(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 375) }
func TestTokenizerEvents_376(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 376) }
func TestTokenizerEvents_377(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 377) }
func TestTokenizerEvents_378(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 378) }
func TestTokenizerEvents_379(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 379) }
func TestTokenizerEvents_380(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 380) }
func TestTokenizerEvents_381(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 381) }
func TestTokenizerEvents_382(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 382) }
func TestTokenizerEvents_383(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 383) }
func TestTokenizerEvents_384(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 384) }
func TestTokenizerEvents_385(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 385) }
func TestTokenizerEvents_386(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 386) }
func TestTokenizerEvents_387(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 387) }
func TestTokenizerEvents_388(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 388) }
func TestTokenizerEvents_389(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 389) }
func TestTokenizerEvents_390(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 390) }
func TestTokenizerEvents_391(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 391) }
func TestTokenizerEvents_392(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 392) }
func TestTokenizerEvents_393(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 393) }
func TestTokenizerEvents_394(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 394) }
func TestTokenizerEvents_395(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 395) }
func TestTokenizerEvents_396(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 396) }
func TestTokenizerEvents_397(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 397) }
func TestTokenizerEvents_398(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 398) }
func TestTokenizerEvents_399(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 399) }
func TestTokenizerEvents_400(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 400) }
func TestTokenizerEvents_401(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 401) }
func TestTokenizerEvents_402(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 402) }
func TestTokenizerEvents_403(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 403) }
func TestTokenizerEvents_404(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 404) }
func TestTokenizerEvents_405(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 405) }
func TestTokenizerEvents_406(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 406) }
func TestTokenizerEvents_407(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 407) }
func TestTokenizerEvents_408(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 408) }
func TestTokenizerEvents_409(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 409) }
func TestTokenizerEvents_410(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 410) }
func TestTokenizerEvents_411(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 411) }
func TestTokenizerEvents_412(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 412) }
func TestTokenizerEvents_413(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 413) }
func TestTokenizerEvents_414(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 414) }
func TestTokenizerEvents_415(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 415) }
func TestTokenizerEvents_416(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 416) }
func TestTokenizerEvents_417(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 417) }
func TestTokenizerEvents_418(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 418) }
func TestTokenizerEvents_419(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 419) }
func TestTokenizerEvents_420(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 420) }
func TestTokenizerEvents_421(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 421) }
func TestTokenizerEvents_422(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 422) }
func TestTokenizerEvents_423(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 423) }
func TestTokenizerEvents_424(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 424) }
func TestTokenizerEvents_425(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 425) }
func TestTokenizerEvents_426(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 426) }
func TestTokenizerEvents_427(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 427) }
func TestTokenizerEvents_428(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 428) }
func TestTokenizerEvents_429(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 429) }
func TestTokenizerEvents_430(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 430) }
func TestTokenizerEvents_431(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 431) }
func TestTokenizerEvents_432(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 432) }
func TestTokenizerEvents_433(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 433) }
func TestTokenizerEvents_434(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 434) }
func TestTokenizerEvents_435(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 435) }
func TestTokenizerEvents_436(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 436) }
func TestTokenizerEvents_437(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 437) }
func TestTokenizerEvents_438(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 438) }
func TestTokenizerEvents_439(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 439) }
func TestTokenizerEvents_440(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 440) }
func TestTokenizerEvents_441(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 441) }
func TestTokenizerEvents_442(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 442) }
func TestTokenizerEvents_443(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 443) }
func TestTokenizerEvents_444(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 444) }
func TestTokenizerEvents_445(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 445) }
func TestTokenizerEvents_446(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 446) }
func TestTokenizerEvents_447(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 447) }
func TestTokenizerEvents_448(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 448) }
func TestTokenizerEvents_449(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 449) }
func TestTokenizerEvents_450(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 450) }
func TestTokenizerEvents_451(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 451) }
func TestTokenizerEvents_452(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 452) }
func TestTokenizerEvents_453(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 453) }
func TestTokenizerEvents_454(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 454) }
func TestTokenizerEvents_455(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 455) }
func TestTokenizerEvents_456(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 456) }
func TestTokenizerEvents_457(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 457) }
func TestTokenizerEvents_458(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 458) }
func TestTokenizerEvents_459(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 459) }
func TestTokenizerEvents_460(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 460) }
func TestTokenizerEvents_461(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 461) }
func TestTokenizerEvents_462(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 462) }
func TestTokenizerEvents_463(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 463) }
func TestTokenizerEvents_464(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 464) }
func TestTokenizerEvents_465(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 465) }
func TestTokenizerEvents_466(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 466) }
func TestTokenizerEvents_467(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 467) }
func TestTokenizerEvents_468(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 468) }
func TestTokenizerEvents_469(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 469) }
func TestTokenizerEvents_470(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 470) }
func TestTokenizerEvents_471(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 471) }
func TestTokenizerEvents_472(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 472) }
func TestTokenizerEvents_473(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 473) }
func TestTokenizerEvents_474(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 474) }
func TestTokenizerEvents_475(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 475) }
func TestTokenizerEvents_476(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 476) }
func TestTokenizerEvents_477(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 477) }
func TestTokenizerEvents_478(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 478) }
func TestTokenizerEvents_479(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 479) }
func TestTokenizerEvents_480(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 480) }
func TestTokenizerEvents_481(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 481) }
func TestTokenizerEvents_482(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 482) }
func TestTokenizerEvents_483(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 483) }
func TestTokenizerEvents_484(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 484) }
func TestTokenizerEvents_485(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 485) }
func TestTokenizerEvents_486(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 486) }
func TestTokenizerEvents_487(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 487) }
func TestTokenizerEvents_488(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 488) }
func TestTokenizerEvents_489(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 489) }
func TestTokenizerEvents_490(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 490) }
func TestTokenizerEvents_491(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 491) }
func TestTokenizerEvents_492(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 492) }
func TestTokenizerEvents_493(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 493) }
func TestTokenizerEvents_494(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 494) }
func TestTokenizerEvents_495(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 495) }
func TestTokenizerEvents_496(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 496) }
func TestTokenizerEvents_497(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 497) }
func TestTokenizerEvents_498(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 498) }
func TestTokenizerEvents_499(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 499) }
func TestTokenizerEvents_500(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 500) }
func TestTokenizerEvents_501(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 501) }
func TestTokenizerEvents_502(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 502) }
func TestTokenizerEvents_503(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 503) }
func TestTokenizerEvents_504(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 504) }
func TestTokenizerEvents_505(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 505) }
func TestTokenizerEvents_506(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 506) }
func TestTokenizerEvents_507(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 507) }
func TestTokenizerEvents_508(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 508) }
func TestTokenizerEvents_509(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 509) }
func TestTokenizerEvents_510(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 510) }
func TestTokenizerEvents_511(t *testing.T) { t.Parallel(); tokenizerEventTopLevelUnit(t, 511) }

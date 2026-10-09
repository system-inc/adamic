// Code generated for the fixed top-level shard table; DO NOT EDIT individual wrappers.
package css

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

var cssParserTopSetup struct {
	once  sync.Once
	units []cssParserShard
	run   func(*testing.T, cssParserShard)
}
var cssParserTopRootOnce sync.Once
var cssParserTopRoot string
var cssParserTopRootError error

func cssParserTopTempDir(t *testing.T) string {
	t.Helper()
	cssParserTopRootOnce.Do(func() { cssParserTopRoot, cssParserTopRootError = os.MkdirTemp("", "adamic-css-top-") })
	if cssParserTopRootError != nil {
		t.Fatal(cssParserTopRootError)
	}
	dir, err := os.MkdirTemp(cssParserTopRoot, "input-")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

// Shared build inputs outlive every parallel leaf and are removed after m.Run.
func TestMain(m *testing.M) {
	code := m.Run()
	if cssParserTopRoot != "" {
		if err := os.RemoveAll(cssParserTopRoot); err != nil {
			if code == 0 {
				code = 1
			}
		}
	}
	os.Exit(code)
}

func cssParserTopPortDirectory(t *testing.T, applied *mutant) string {
	t.Helper()
	original := portDirectory(t, applied)
	destination := filepath.Join(cssParserTopTempDir(t), "sources")
	if err := os.Rename(filepath.Dir(original), destination); err != nil {
		t.Fatal(err)
	}
	return filepath.Join(destination, "css")
}

func TestThePortParsesAsGoCohereDoes_000(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 0)
}
func TestThePortParsesAsGoCohereDoes_001(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 1)
}
func TestThePortParsesAsGoCohereDoes_002(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 2)
}
func TestThePortParsesAsGoCohereDoes_003(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 3)
}
func TestThePortParsesAsGoCohereDoes_004(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 4)
}
func TestThePortParsesAsGoCohereDoes_005(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 5)
}
func TestThePortParsesAsGoCohereDoes_006(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 6)
}
func TestThePortParsesAsGoCohereDoes_007(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 7)
}
func TestThePortParsesAsGoCohereDoes_008(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 8)
}
func TestThePortParsesAsGoCohereDoes_009(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 9)
}
func TestThePortParsesAsGoCohereDoes_010(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 10)
}
func TestThePortParsesAsGoCohereDoes_011(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 11)
}
func TestThePortParsesAsGoCohereDoes_012(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 12)
}
func TestThePortParsesAsGoCohereDoes_013(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 13)
}
func TestThePortParsesAsGoCohereDoes_014(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 14)
}
func TestThePortParsesAsGoCohereDoes_015(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 15)
}
func TestThePortParsesAsGoCohereDoes_016(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 16)
}
func TestThePortParsesAsGoCohereDoes_017(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 17)
}
func TestThePortParsesAsGoCohereDoes_018(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 18)
}
func TestThePortParsesAsGoCohereDoes_019(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 19)
}
func TestThePortParsesAsGoCohereDoes_020(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 20)
}
func TestThePortParsesAsGoCohereDoes_021(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 21)
}
func TestThePortParsesAsGoCohereDoes_022(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 22)
}
func TestThePortParsesAsGoCohereDoes_023(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 23)
}
func TestThePortParsesAsGoCohereDoes_024(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 24)
}
func TestThePortParsesAsGoCohereDoes_025(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 25)
}
func TestThePortParsesAsGoCohereDoes_026(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 26)
}
func TestThePortParsesAsGoCohereDoes_027(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 27)
}
func TestThePortParsesAsGoCohereDoes_028(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 28)
}
func TestThePortParsesAsGoCohereDoes_029(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 29)
}
func TestThePortParsesAsGoCohereDoes_030(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 30)
}
func TestThePortParsesAsGoCohereDoes_031(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 31)
}
func TestThePortParsesAsGoCohereDoes_032(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 32)
}
func TestThePortParsesAsGoCohereDoes_033(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 33)
}
func TestThePortParsesAsGoCohereDoes_034(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 34)
}
func TestThePortParsesAsGoCohereDoes_035(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 35)
}
func TestThePortParsesAsGoCohereDoes_036(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 36)
}
func TestThePortParsesAsGoCohereDoes_037(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 37)
}
func TestThePortParsesAsGoCohereDoes_038(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 38)
}
func TestThePortParsesAsGoCohereDoes_039(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 39)
}
func TestThePortParsesAsGoCohereDoes_040(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 40)
}
func TestThePortParsesAsGoCohereDoes_041(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 41)
}
func TestThePortParsesAsGoCohereDoes_042(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 42)
}
func TestThePortParsesAsGoCohereDoes_043(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 43)
}
func TestThePortParsesAsGoCohereDoes_044(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 44)
}
func TestThePortParsesAsGoCohereDoes_045(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 45)
}
func TestThePortParsesAsGoCohereDoes_046(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 46)
}
func TestThePortParsesAsGoCohereDoes_047(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 47)
}
func TestThePortParsesAsGoCohereDoes_048(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 48)
}
func TestThePortParsesAsGoCohereDoes_049(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 49)
}
func TestThePortParsesAsGoCohereDoes_050(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 50)
}
func TestThePortParsesAsGoCohereDoes_051(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 51)
}
func TestThePortParsesAsGoCohereDoes_052(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 52)
}
func TestThePortParsesAsGoCohereDoes_053(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 53)
}
func TestThePortParsesAsGoCohereDoes_054(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 54)
}
func TestThePortParsesAsGoCohereDoes_055(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 55)
}
func TestThePortParsesAsGoCohereDoes_056(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 56)
}
func TestThePortParsesAsGoCohereDoes_057(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 57)
}
func TestThePortParsesAsGoCohereDoes_058(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 58)
}
func TestThePortParsesAsGoCohereDoes_059(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 59)
}
func TestThePortParsesAsGoCohereDoes_060(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 60)
}
func TestThePortParsesAsGoCohereDoes_061(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 61)
}
func TestThePortParsesAsGoCohereDoes_062(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 62)
}
func TestThePortParsesAsGoCohereDoes_063(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 63)
}
func TestThePortParsesAsGoCohereDoes_064(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 64)
}
func TestThePortParsesAsGoCohereDoes_065(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 65)
}
func TestThePortParsesAsGoCohereDoes_066(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 66)
}
func TestThePortParsesAsGoCohereDoes_067(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 67)
}
func TestThePortParsesAsGoCohereDoes_068(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 68)
}
func TestThePortParsesAsGoCohereDoes_069(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 69)
}
func TestThePortParsesAsGoCohereDoes_070(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 70)
}
func TestThePortParsesAsGoCohereDoes_071(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 71)
}
func TestThePortParsesAsGoCohereDoes_072(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 72)
}
func TestThePortParsesAsGoCohereDoes_073(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 73)
}
func TestThePortParsesAsGoCohereDoes_074(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 74)
}
func TestThePortParsesAsGoCohereDoes_075(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 75)
}
func TestThePortParsesAsGoCohereDoes_076(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 76)
}
func TestThePortParsesAsGoCohereDoes_077(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 77)
}
func TestThePortParsesAsGoCohereDoes_078(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 78)
}
func TestThePortParsesAsGoCohereDoes_079(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 79)
}
func TestThePortParsesAsGoCohereDoes_080(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 80)
}
func TestThePortParsesAsGoCohereDoes_081(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 81)
}
func TestThePortParsesAsGoCohereDoes_082(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 82)
}
func TestThePortParsesAsGoCohereDoes_083(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 83)
}
func TestThePortParsesAsGoCohereDoes_084(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 84)
}
func TestThePortParsesAsGoCohereDoes_085(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 85)
}
func TestThePortParsesAsGoCohereDoes_086(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 86)
}
func TestThePortParsesAsGoCohereDoes_087(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 87)
}
func TestThePortParsesAsGoCohereDoes_088(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 88)
}
func TestThePortParsesAsGoCohereDoes_089(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 89)
}
func TestThePortParsesAsGoCohereDoes_090(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 90)
}
func TestThePortParsesAsGoCohereDoes_091(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 91)
}
func TestThePortParsesAsGoCohereDoes_092(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 92)
}
func TestThePortParsesAsGoCohereDoes_093(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 93)
}
func TestThePortParsesAsGoCohereDoes_094(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 94)
}
func TestThePortParsesAsGoCohereDoes_095(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 95)
}
func TestThePortParsesAsGoCohereDoes_096(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 96)
}
func TestThePortParsesAsGoCohereDoes_097(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 97)
}
func TestThePortParsesAsGoCohereDoes_098(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 98)
}
func TestThePortParsesAsGoCohereDoes_099(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 99)
}
func TestThePortParsesAsGoCohereDoes_100(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 100)
}
func TestThePortParsesAsGoCohereDoes_101(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 101)
}
func TestThePortParsesAsGoCohereDoes_102(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 102)
}
func TestThePortParsesAsGoCohereDoes_103(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 103)
}
func TestThePortParsesAsGoCohereDoes_104(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 104)
}
func TestThePortParsesAsGoCohereDoes_105(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 105)
}
func TestThePortParsesAsGoCohereDoes_106(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 106)
}
func TestThePortParsesAsGoCohereDoes_107(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 107)
}
func TestThePortParsesAsGoCohereDoes_108(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 108)
}
func TestThePortParsesAsGoCohereDoes_109(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 109)
}
func TestThePortParsesAsGoCohereDoes_110(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 110)
}
func TestThePortParsesAsGoCohereDoes_111(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 111)
}
func TestThePortParsesAsGoCohereDoes_112(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 112)
}
func TestThePortParsesAsGoCohereDoes_113(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 113)
}
func TestThePortParsesAsGoCohereDoes_114(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 114)
}
func TestThePortParsesAsGoCohereDoes_115(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 115)
}
func TestThePortParsesAsGoCohereDoes_116(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 116)
}
func TestThePortParsesAsGoCohereDoes_117(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 117)
}
func TestThePortParsesAsGoCohereDoes_118(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 118)
}
func TestThePortParsesAsGoCohereDoes_119(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 119)
}
func TestThePortParsesAsGoCohereDoes_120(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 120)
}
func TestThePortParsesAsGoCohereDoes_121(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 121)
}
func TestThePortParsesAsGoCohereDoes_122(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 122)
}
func TestThePortParsesAsGoCohereDoes_123(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 123)
}
func TestThePortParsesAsGoCohereDoes_124(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 124)
}
func TestThePortParsesAsGoCohereDoes_125(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 125)
}
func TestThePortParsesAsGoCohereDoes_126(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 126)
}
func TestThePortParsesAsGoCohereDoes_127(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 127)
}
func TestThePortParsesAsGoCohereDoes_128(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 128)
}
func TestThePortParsesAsGoCohereDoes_129(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 129)
}
func TestThePortParsesAsGoCohereDoes_130(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 130)
}
func TestThePortParsesAsGoCohereDoes_131(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 131)
}
func TestThePortParsesAsGoCohereDoes_132(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 132)
}
func TestThePortParsesAsGoCohereDoes_133(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 133)
}
func TestThePortParsesAsGoCohereDoes_134(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 134)
}
func TestThePortParsesAsGoCohereDoes_135(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 135)
}
func TestThePortParsesAsGoCohereDoes_136(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 136)
}
func TestThePortParsesAsGoCohereDoes_137(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 137)
}
func TestThePortParsesAsGoCohereDoes_138(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 138)
}
func TestThePortParsesAsGoCohereDoes_139(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 139)
}
func TestThePortParsesAsGoCohereDoes_140(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 140)
}
func TestThePortParsesAsGoCohereDoes_141(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 141)
}
func TestThePortParsesAsGoCohereDoes_142(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 142)
}
func TestThePortParsesAsGoCohereDoes_143(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 143)
}
func TestThePortParsesAsGoCohereDoes_144(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 144)
}
func TestThePortParsesAsGoCohereDoes_145(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 145)
}
func TestThePortParsesAsGoCohereDoes_146(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 146)
}
func TestThePortParsesAsGoCohereDoes_147(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 147)
}
func TestThePortParsesAsGoCohereDoes_148(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 148)
}
func TestThePortParsesAsGoCohereDoes_149(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 149)
}
func TestThePortParsesAsGoCohereDoes_150(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 150)
}
func TestThePortParsesAsGoCohereDoes_151(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 151)
}
func TestThePortParsesAsGoCohereDoes_152(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 152)
}
func TestThePortParsesAsGoCohereDoes_153(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 153)
}
func TestThePortParsesAsGoCohereDoes_154(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 154)
}
func TestThePortParsesAsGoCohereDoes_155(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 155)
}
func TestThePortParsesAsGoCohereDoes_156(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 156)
}
func TestThePortParsesAsGoCohereDoes_157(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 157)
}
func TestThePortParsesAsGoCohereDoes_158(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 158)
}
func TestThePortParsesAsGoCohereDoes_159(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 159)
}
func TestThePortParsesAsGoCohereDoes_160(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 160)
}
func TestThePortParsesAsGoCohereDoes_161(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 161)
}
func TestThePortParsesAsGoCohereDoes_162(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 162)
}
func TestThePortParsesAsGoCohereDoes_163(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 163)
}
func TestThePortParsesAsGoCohereDoes_164(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 164)
}
func TestThePortParsesAsGoCohereDoes_165(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 165)
}
func TestThePortParsesAsGoCohereDoes_166(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 166)
}
func TestThePortParsesAsGoCohereDoes_167(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 167)
}
func TestThePortParsesAsGoCohereDoes_168(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 168)
}
func TestThePortParsesAsGoCohereDoes_169(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 169)
}
func TestThePortParsesAsGoCohereDoes_170(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 170)
}
func TestThePortParsesAsGoCohereDoes_171(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 171)
}
func TestThePortParsesAsGoCohereDoes_172(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 172)
}
func TestThePortParsesAsGoCohereDoes_173(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 173)
}
func TestThePortParsesAsGoCohereDoes_174(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 174)
}
func TestThePortParsesAsGoCohereDoes_175(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 175)
}
func TestThePortParsesAsGoCohereDoes_176(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 176)
}
func TestThePortParsesAsGoCohereDoes_177(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 177)
}
func TestThePortParsesAsGoCohereDoes_178(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 178)
}
func TestThePortParsesAsGoCohereDoes_179(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 179)
}
func TestThePortParsesAsGoCohereDoes_180(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 180)
}
func TestThePortParsesAsGoCohereDoes_181(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 181)
}
func TestThePortParsesAsGoCohereDoes_182(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 182)
}
func TestThePortParsesAsGoCohereDoes_183(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 183)
}
func TestThePortParsesAsGoCohereDoes_184(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 184)
}
func TestThePortParsesAsGoCohereDoes_185(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 185)
}
func TestThePortParsesAsGoCohereDoes_186(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 186)
}
func TestThePortParsesAsGoCohereDoes_187(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 187)
}
func TestThePortParsesAsGoCohereDoes_188(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 188)
}
func TestThePortParsesAsGoCohereDoes_189(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 189)
}
func TestThePortParsesAsGoCohereDoes_190(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 190)
}
func TestThePortParsesAsGoCohereDoes_191(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 191)
}
func TestThePortParsesAsGoCohereDoes_192(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 192)
}
func TestThePortParsesAsGoCohereDoes_193(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 193)
}
func TestThePortParsesAsGoCohereDoes_194(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 194)
}
func TestThePortParsesAsGoCohereDoes_195(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 195)
}
func TestThePortParsesAsGoCohereDoes_196(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 196)
}
func TestThePortParsesAsGoCohereDoes_197(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 197)
}
func TestThePortParsesAsGoCohereDoes_198(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 198)
}
func TestThePortParsesAsGoCohereDoes_199(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 199)
}
func TestThePortParsesAsGoCohereDoes_200(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 200)
}
func TestThePortParsesAsGoCohereDoes_201(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 201)
}
func TestThePortParsesAsGoCohereDoes_202(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 202)
}
func TestThePortParsesAsGoCohereDoes_203(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 203)
}
func TestThePortParsesAsGoCohereDoes_204(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 204)
}
func TestThePortParsesAsGoCohereDoes_205(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 205)
}
func TestThePortParsesAsGoCohereDoes_206(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 206)
}
func TestThePortParsesAsGoCohereDoes_207(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 207)
}
func TestThePortParsesAsGoCohereDoes_208(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 208)
}
func TestThePortParsesAsGoCohereDoes_209(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 209)
}
func TestThePortParsesAsGoCohereDoes_210(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 210)
}
func TestThePortParsesAsGoCohereDoes_211(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 211)
}
func TestThePortParsesAsGoCohereDoes_212(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 212)
}
func TestThePortParsesAsGoCohereDoes_213(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 213)
}
func TestThePortParsesAsGoCohereDoes_214(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 214)
}
func TestThePortParsesAsGoCohereDoes_215(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 215)
}
func TestThePortParsesAsGoCohereDoes_216(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 216)
}
func TestThePortParsesAsGoCohereDoes_217(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 217)
}
func TestThePortParsesAsGoCohereDoes_218(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 218)
}
func TestThePortParsesAsGoCohereDoes_219(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 219)
}
func TestThePortParsesAsGoCohereDoes_220(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 220)
}
func TestThePortParsesAsGoCohereDoes_221(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 221)
}
func TestThePortParsesAsGoCohereDoes_222(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 222)
}
func TestThePortParsesAsGoCohereDoes_223(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 223)
}
func TestThePortParsesAsGoCohereDoes_224(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 224)
}
func TestThePortParsesAsGoCohereDoes_225(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 225)
}
func TestThePortParsesAsGoCohereDoes_226(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 226)
}
func TestThePortParsesAsGoCohereDoes_227(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 227)
}
func TestThePortParsesAsGoCohereDoes_228(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 228)
}
func TestThePortParsesAsGoCohereDoes_229(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 229)
}
func TestThePortParsesAsGoCohereDoes_230(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 230)
}
func TestThePortParsesAsGoCohereDoes_231(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 231)
}
func TestThePortParsesAsGoCohereDoes_232(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 232)
}
func TestThePortParsesAsGoCohereDoes_233(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 233)
}
func TestThePortParsesAsGoCohereDoes_234(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 234)
}
func TestThePortParsesAsGoCohereDoes_235(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 235)
}
func TestThePortParsesAsGoCohereDoes_236(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 236)
}
func TestThePortParsesAsGoCohereDoes_237(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 237)
}
func TestThePortParsesAsGoCohereDoes_238(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 238)
}
func TestThePortParsesAsGoCohereDoes_239(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 239)
}
func TestThePortParsesAsGoCohereDoes_240(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 240)
}
func TestThePortParsesAsGoCohereDoes_241(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 241)
}
func TestThePortParsesAsGoCohereDoes_242(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 242)
}
func TestThePortParsesAsGoCohereDoes_243(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 243)
}
func TestThePortParsesAsGoCohereDoes_244(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 244)
}
func TestThePortParsesAsGoCohereDoes_245(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 245)
}
func TestThePortParsesAsGoCohereDoes_246(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 246)
}
func TestThePortParsesAsGoCohereDoes_247(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 247)
}
func TestThePortParsesAsGoCohereDoes_248(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 248)
}
func TestThePortParsesAsGoCohereDoes_249(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 249)
}
func TestThePortParsesAsGoCohereDoes_250(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 250)
}
func TestThePortParsesAsGoCohereDoes_251(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 251)
}
func TestThePortParsesAsGoCohereDoes_252(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 252)
}
func TestThePortParsesAsGoCohereDoes_253(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 253)
}
func TestThePortParsesAsGoCohereDoes_254(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 254)
}
func TestThePortParsesAsGoCohereDoes_255(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 255)
}
func TestThePortParsesAsGoCohereDoes_256(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 256)
}
func TestThePortParsesAsGoCohereDoes_257(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 257)
}
func TestThePortParsesAsGoCohereDoes_258(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 258)
}
func TestThePortParsesAsGoCohereDoes_259(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 259)
}
func TestThePortParsesAsGoCohereDoes_260(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 260)
}
func TestThePortParsesAsGoCohereDoes_261(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 261)
}
func TestThePortParsesAsGoCohereDoes_262(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 262)
}
func TestThePortParsesAsGoCohereDoes_263(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 263)
}
func TestThePortParsesAsGoCohereDoes_264(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 264)
}
func TestThePortParsesAsGoCohereDoes_265(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 265)
}
func TestThePortParsesAsGoCohereDoes_266(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 266)
}
func TestThePortParsesAsGoCohereDoes_267(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 267)
}
func TestThePortParsesAsGoCohereDoes_268(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 268)
}
func TestThePortParsesAsGoCohereDoes_269(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 269)
}
func TestThePortParsesAsGoCohereDoes_270(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 270)
}
func TestThePortParsesAsGoCohereDoes_271(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 271)
}
func TestThePortParsesAsGoCohereDoes_272(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 272)
}
func TestThePortParsesAsGoCohereDoes_273(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 273)
}
func TestThePortParsesAsGoCohereDoes_274(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 274)
}
func TestThePortParsesAsGoCohereDoes_275(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 275)
}
func TestThePortParsesAsGoCohereDoes_276(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 276)
}
func TestThePortParsesAsGoCohereDoes_277(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 277)
}
func TestThePortParsesAsGoCohereDoes_278(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 278)
}
func TestThePortParsesAsGoCohereDoes_279(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 279)
}
func TestThePortParsesAsGoCohereDoes_280(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 280)
}
func TestThePortParsesAsGoCohereDoes_281(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 281)
}
func TestThePortParsesAsGoCohereDoes_282(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 282)
}
func TestThePortParsesAsGoCohereDoes_283(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 283)
}
func TestThePortParsesAsGoCohereDoes_284(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 284)
}
func TestThePortParsesAsGoCohereDoes_285(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 285)
}
func TestThePortParsesAsGoCohereDoes_286(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 286)
}
func TestThePortParsesAsGoCohereDoes_287(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 287)
}
func TestThePortParsesAsGoCohereDoes_288(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 288)
}
func TestThePortParsesAsGoCohereDoes_289(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 289)
}
func TestThePortParsesAsGoCohereDoes_290(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 290)
}
func TestThePortParsesAsGoCohereDoes_291(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 291)
}
func TestThePortParsesAsGoCohereDoes_292(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 292)
}
func TestThePortParsesAsGoCohereDoes_293(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 293)
}
func TestThePortParsesAsGoCohereDoes_294(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 294)
}
func TestThePortParsesAsGoCohereDoes_295(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 295)
}
func TestThePortParsesAsGoCohereDoes_296(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 296)
}
func TestThePortParsesAsGoCohereDoes_297(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 297)
}
func TestThePortParsesAsGoCohereDoes_298(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 298)
}
func TestThePortParsesAsGoCohereDoes_299(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 299)
}
func TestThePortParsesAsGoCohereDoes_300(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 300)
}
func TestThePortParsesAsGoCohereDoes_301(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 301)
}
func TestThePortParsesAsGoCohereDoes_302(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 302)
}
func TestThePortParsesAsGoCohereDoes_303(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 303)
}
func TestThePortParsesAsGoCohereDoes_304(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 304)
}
func TestThePortParsesAsGoCohereDoes_305(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 305)
}
func TestThePortParsesAsGoCohereDoes_306(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 306)
}
func TestThePortParsesAsGoCohereDoes_307(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 307)
}
func TestThePortParsesAsGoCohereDoes_308(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 308)
}
func TestThePortParsesAsGoCohereDoes_309(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 309)
}
func TestThePortParsesAsGoCohereDoes_310(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 310)
}
func TestThePortParsesAsGoCohereDoes_311(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 311)
}
func TestThePortParsesAsGoCohereDoes_312(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 312)
}
func TestThePortParsesAsGoCohereDoes_313(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 313)
}
func TestThePortParsesAsGoCohereDoes_314(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 314)
}
func TestThePortParsesAsGoCohereDoes_315(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 315)
}
func TestThePortParsesAsGoCohereDoes_316(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 316)
}
func TestThePortParsesAsGoCohereDoes_317(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 317)
}
func TestThePortParsesAsGoCohereDoes_318(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 318)
}
func TestThePortParsesAsGoCohereDoes_319(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 319)
}
func TestThePortParsesAsGoCohereDoes_320(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 320)
}
func TestThePortParsesAsGoCohereDoes_321(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 321)
}
func TestThePortParsesAsGoCohereDoes_322(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 322)
}
func TestThePortParsesAsGoCohereDoes_323(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 323)
}
func TestThePortParsesAsGoCohereDoes_324(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 324)
}
func TestThePortParsesAsGoCohereDoes_325(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 325)
}
func TestThePortParsesAsGoCohereDoes_326(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 326)
}
func TestThePortParsesAsGoCohereDoes_327(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 327)
}
func TestThePortParsesAsGoCohereDoes_328(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 328)
}
func TestThePortParsesAsGoCohereDoes_329(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 329)
}
func TestThePortParsesAsGoCohereDoes_330(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 330)
}
func TestThePortParsesAsGoCohereDoes_331(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 331)
}
func TestThePortParsesAsGoCohereDoes_332(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 332)
}
func TestThePortParsesAsGoCohereDoes_333(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 333)
}
func TestThePortParsesAsGoCohereDoes_334(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 334)
}
func TestThePortParsesAsGoCohereDoes_335(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 335)
}
func TestThePortParsesAsGoCohereDoes_336(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 336)
}
func TestThePortParsesAsGoCohereDoes_337(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 337)
}
func TestThePortParsesAsGoCohereDoes_338(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 338)
}
func TestThePortParsesAsGoCohereDoes_339(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 339)
}
func TestThePortParsesAsGoCohereDoes_340(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 340)
}
func TestThePortParsesAsGoCohereDoes_341(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 341)
}
func TestThePortParsesAsGoCohereDoes_342(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 342)
}
func TestThePortParsesAsGoCohereDoes_343(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 343)
}
func TestThePortParsesAsGoCohereDoes_344(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 344)
}
func TestThePortParsesAsGoCohereDoes_345(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 345)
}
func TestThePortParsesAsGoCohereDoes_346(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 346)
}
func TestThePortParsesAsGoCohereDoes_347(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 347)
}
func TestThePortParsesAsGoCohereDoes_348(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 348)
}
func TestThePortParsesAsGoCohereDoes_349(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 349)
}
func TestThePortParsesAsGoCohereDoes_350(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 350)
}
func TestThePortParsesAsGoCohereDoes_351(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 351)
}
func TestThePortParsesAsGoCohereDoes_352(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 352)
}
func TestThePortParsesAsGoCohereDoes_353(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 353)
}
func TestThePortParsesAsGoCohereDoes_354(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 354)
}
func TestThePortParsesAsGoCohereDoes_355(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 355)
}
func TestThePortParsesAsGoCohereDoes_356(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 356)
}
func TestThePortParsesAsGoCohereDoes_357(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 357)
}
func TestThePortParsesAsGoCohereDoes_358(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 358)
}
func TestThePortParsesAsGoCohereDoes_359(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 359)
}
func TestThePortParsesAsGoCohereDoes_360(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 360)
}
func TestThePortParsesAsGoCohereDoes_361(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 361)
}
func TestThePortParsesAsGoCohereDoes_362(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 362)
}
func TestThePortParsesAsGoCohereDoes_363(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 363)
}
func TestThePortParsesAsGoCohereDoes_364(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 364)
}
func TestThePortParsesAsGoCohereDoes_365(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 365)
}
func TestThePortParsesAsGoCohereDoes_366(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 366)
}
func TestThePortParsesAsGoCohereDoes_367(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 367)
}
func TestThePortParsesAsGoCohereDoes_368(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 368)
}
func TestThePortParsesAsGoCohereDoes_369(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 369)
}
func TestThePortParsesAsGoCohereDoes_370(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 370)
}
func TestThePortParsesAsGoCohereDoes_371(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 371)
}
func TestThePortParsesAsGoCohereDoes_372(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 372)
}
func TestThePortParsesAsGoCohereDoes_373(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 373)
}
func TestThePortParsesAsGoCohereDoes_374(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 374)
}
func TestThePortParsesAsGoCohereDoes_375(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 375)
}
func TestThePortParsesAsGoCohereDoes_376(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 376)
}
func TestThePortParsesAsGoCohereDoes_377(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 377)
}
func TestThePortParsesAsGoCohereDoes_378(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 378)
}
func TestThePortParsesAsGoCohereDoes_379(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 379)
}
func TestThePortParsesAsGoCohereDoes_380(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 380)
}
func TestThePortParsesAsGoCohereDoes_381(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 381)
}
func TestThePortParsesAsGoCohereDoes_382(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 382)
}
func TestThePortParsesAsGoCohereDoes_383(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 383)
}
func TestThePortParsesAsGoCohereDoes_384(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 384)
}
func TestThePortParsesAsGoCohereDoes_385(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 385)
}
func TestThePortParsesAsGoCohereDoes_386(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 386)
}
func TestThePortParsesAsGoCohereDoes_387(t *testing.T) {
	t.Parallel()
	runThePortParsesAsGoCohereDoesShard(t, 387)
}

var cssParserTopCorpusSetup struct {
	once   sync.Once
	corpus cssParserCorpus
	units  []cssParserShard
}

func cssParserTopCorpus(t *testing.T) cssParserCorpus {
	t.Helper()
	cssParserTopCorpusSetup.once.Do(func() {
		oracle := cssParserOracle(t)
		cases, answers := cachedCSSParserOracleOutputs(t, oracle)
		cssParserTopCorpusSetup.corpus = readCSSParserCorpus(t, cases, answers)
		cssParserTopCorpusSetup.units = cssParserShards(cssParserTopCorpusSetup.corpus.keys)
	})
	if len(cssParserTopCorpusSetup.corpus.keys) == 0 {
		t.Fatal("shared corpus setup did not complete")
	}
	return cssParserTopCorpusSetup.corpus
}
func cssParserContainsCase(indices []int, witness int) bool {
	for _, i := range indices {
		if i == witness {
			return true
		}
	}
	return false
}

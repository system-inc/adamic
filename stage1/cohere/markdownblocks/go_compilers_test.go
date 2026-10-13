package markdownblocks

import (
	"maps"
	"slices"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// markdownGoBuild is one Go build's key, by the compilers go builds it with.
type markdownGoBuild struct {
	name   string
	inputs func(markdownCompilers) buildcache.Inputs
}

// markdownGoBuilds are every Go build in the package whose key names the compilers: the text builds and each grain
// recipe's Go oracle.
func markdownGoBuilds() []markdownGoBuild {
	var builds []markdownGoBuild
	for _, target := range []string{"go", "formatter", "generator"} {
		builds = append(builds, markdownGoBuild{"text " + target, func(compilers markdownCompilers) buildcache.Inputs { return textGoProductInputs(target, compilers) }})
	}
	for _, name := range slices.Sorted(maps.Keys(grainRecipes)) {
		builds = append(builds, markdownGoBuild{"grain " + name, func(compilers markdownCompilers) buildcache.Inputs { return grainGoInputs(name, compilers) }})
	}
	return builds
}

// Each Go build's key moves with the C or C++ compiler go builds it with, and with the C compiler's report, so a
// changed compiler misses rather than reading a product another one built.
func TestMarkdownGoBuildsRekeyCompiler(t *testing.T) {
	t.Parallel()
	root := markdownLayoutProductRoot(t)
	compilers := markdownCompilers{"clang", "clang++", "clang version 18.1.8"}
	for _, build := range markdownGoBuilds() {
		key := buildcache.NameKey(root, build.inputs(compilers))
		for _, other := range []markdownCompilers{{"gcc", "clang++", compilers.report}, {"clang", "g++", compilers.report}, {"clang", "clang++", "clang version 19.1.7"}} {
			if otherKey := buildcache.NameKey(root, build.inputs(other)); otherKey == key {
				t.Fatalf("%s: compilers %+v kept its key %s", build.name, other, key[:12])
			}
		}
		t.Logf("%s key %s moves with CC, CXX and the C compiler's report", build.name, key[:12])
	}
}

// The compilers the Go builds' keys name are read from toolchain.Env, the environment go runs under, not from the
// caller's: with CC and CXX unset that is clang and clang++, which a reader of os.Getenv would key as empty while go ran
// clang, and a CC the caller sets moves the key. Not parallel: it sets CC and CXX.
func TestMarkdownGoBuildsKeyTheCompilersEnvGivesGo(t *testing.T) {
	root := markdownLayoutProductRoot(t)
	t.Setenv("CC", "")
	t.Setenv("CXX", "")
	defaults := markdownGoCompilers()
	builds := markdownGoBuilds()
	keys := map[string]string{}
	for _, build := range builds {
		inputs := build.inputs(defaults)
		if !slices.Contains(inputs.Flags, "CC=clang") || !slices.Contains(inputs.Flags, "CXX=clang++") || !slices.Contains(inputs.Toolchain, buildcache.Tool("clang", "--version")) {
			t.Fatalf("with CC and CXX unset, %s's key names %+v, want clang, clang++ and clang's report", build.name, defaults)
		}
		keys[build.name] = buildcache.NameKey(root, inputs)
	}
	t.Setenv("CC", "cc")
	moved := markdownGoCompilers()
	if moved.c != "cc" || moved.report == defaults.report {
		t.Fatalf("CC=cc read as %+v", moved)
	}
	for _, build := range builds {
		if buildcache.NameKey(root, build.inputs(moved)) == keys[build.name] {
			t.Fatalf("CC=cc kept %s's key", build.name)
		}
	}
	t.Logf("unset: %s and %s; CC=cc moves every Go build's key", defaults.c, defaults.cPlusPlus)
}

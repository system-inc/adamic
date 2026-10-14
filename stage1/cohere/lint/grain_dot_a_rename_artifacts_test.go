package lint

import (
	"github.com/system-inc/adamic/internal/buildcache"
	"testing"
	"time"
)

func TestProduct_DotARenameSource(t *testing.T) {
	t.Parallel()
	started := time.Now()
	buildcache.RequireArtifacts(t, dotARenameSource(t), buildcache.Artifact{Name: "main.ts"}, buildcache.Artifact{Name: "rules/no-var/rule.ts"}, buildcache.Artifact{Name: "rules/no-var/rule.json"})
	t.Logf("%s: %.3fs", t.Name(), time.Since(started).Seconds())
}

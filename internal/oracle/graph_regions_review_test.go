package oracle

import "testing"

func init() {
	for _, path := range graphReviewFixtures {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}

var graphReviewFixtures = []string{
	"internal/oracle/testdata/review/agree/cwd.a",
	"internal/oracle/testdata/review/agree/fxspptb_aee89b2_alias_fill.a",
	"internal/oracle/testdata/review/agree/fxspptb_aee89b2_alias_mapset.a",
	"internal/oracle/testdata/review/agree/fxspptb_aee89b2_alias_reverse.a",
	"internal/oracle/testdata/review/agree/fxspptb_aee89b2_alias_setadd.a",
	"internal/oracle/testdata/review/agree/fxspptb_aee89b2_alias_sort.a",
	"internal/oracle/testdata/review/agree/fxspptb_aee89b2_clobber_link_loop.a",
	"internal/oracle/testdata/review/agree/fxspptb_aee89b2_escaped_before.a",
	"internal/oracle/testdata/review/agree/fxspptb_aee89b2_map_entries_pattern.a",
	"internal/oracle/testdata/review/agree/fxspptb_aee89b2_spread_outside.a",
	"internal/oracle/testdata/review/agree/fxspptb_aee89b2_throw_keeps_old.a",
	"internal/oracle/testdata/review/agree/fxspptb_devirt_fresh_method_keeps_argument.a",
}

func TestGraphReviewCwd(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/review/agree/cwd.a")
}

func TestGraphReviewFxspptbAee89B2AliasFill(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/review/agree/fxspptb_aee89b2_alias_fill.a")
}

func TestGraphReviewFxspptbAee89B2AliasMapset(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/review/agree/fxspptb_aee89b2_alias_mapset.a")
}

func TestGraphReviewFxspptbAee89B2AliasReverse(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/review/agree/fxspptb_aee89b2_alias_reverse.a")
}

func TestGraphReviewFxspptbAee89B2AliasSetadd(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/review/agree/fxspptb_aee89b2_alias_setadd.a")
}

func TestGraphReviewFxspptbAee89B2AliasSort(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/review/agree/fxspptb_aee89b2_alias_sort.a")
}

func TestGraphReviewFxspptbAee89B2ClobberLinkLoop(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/review/agree/fxspptb_aee89b2_clobber_link_loop.a")
}

func TestGraphReviewFxspptbAee89B2EscapedBefore(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/review/agree/fxspptb_aee89b2_escaped_before.a")
}

func TestGraphReviewFxspptbAee89B2MapEntriesPattern(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/review/agree/fxspptb_aee89b2_map_entries_pattern.a")
}

func TestGraphReviewFxspptbAee89B2SpreadOutside(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/review/agree/fxspptb_aee89b2_spread_outside.a")
}

func TestGraphReviewFxspptbAee89B2ThrowKeepsOld(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/review/agree/fxspptb_aee89b2_throw_keeps_old.a")
}

func TestGraphReviewFxspptbDevirtFreshMethodKeepsArgument(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/review/agree/fxspptb_devirt_fresh_method_keeps_argument.a")
}

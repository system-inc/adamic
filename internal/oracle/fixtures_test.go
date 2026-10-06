package oracle

import (
	"github.com/system-inc/adamic/internal/oracle/fixturedata"
	"testing"
)

type oracleFixture struct {
	path                                           string
	lowers, checked, unreadable, writes, uncounted bool
	arguments                                      []string
	countsPath, optionsIdentity                    string
}

func oracleFixtures(t *testing.T, input bool) []oracleFixture {
	t.Helper()
	discovered, err := fixturedata.Discover(repository)
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []oracleFixture
	for _, fixture := range discovered {
		if fixture.Input == input {
			fixtures = append(fixtures, oracleFixture{path: fixture.Path, lowers: fixture.Lowers, checked: fixture.Checked,
				arguments: fixture.Arguments, unreadable: fixture.Unreadable, writes: fixture.Writes, uncounted: fixture.Uncounted,
				countsPath: fixture.CountsPath, optionsIdentity: fixture.OptionsIdentity})
		}
	}
	return fixtures
}

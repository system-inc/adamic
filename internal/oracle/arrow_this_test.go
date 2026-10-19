package oracle

import "testing"

func init() {
	for _, name := range []string{"arrow_this_super.a", "arrow_this_field.a", "arrow_this_stored.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/" + name, true, false})
	}
}

func TestArrowThisSuper(t *testing.T) {
	t.Parallel()
	callTargetFixtureAgrees(t, "arrow_this_super.a", "arrow-super\n")
}

func TestArrowThisField(t *testing.T) {
	t.Parallel()
	callTargetFixtureAgrees(t, "arrow_this_field.a", "arrow-field arrow-field\n")
}

func TestArrowThisStored(t *testing.T) {
	t.Parallel()
	callTargetFixtureAgrees(t, "arrow_this_stored.a", "arrow-stored/arrow-stored\n")
}

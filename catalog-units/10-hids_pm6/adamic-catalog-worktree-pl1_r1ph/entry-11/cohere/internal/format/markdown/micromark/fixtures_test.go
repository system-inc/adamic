package micromark

// eventFixtures are compared event for event against upstream by TestEventsMatchUpstream. Each construct's
// port adds the cases that exercise it, including the ones that must fail to match.
var eventFixtures = []string{
	"",
	"a",
	"a\nb",
	"a\n\nb",
	"  a  \n b",
	"***",
	"- - -",
	"_ _ _ _",
	"**",
	"a\\*b\\",
	"a\\\nb",
	"a  \nb",
	"a\t\nb",
	"a \n",
}

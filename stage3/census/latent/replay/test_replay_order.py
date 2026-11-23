"""The replay's entry reach keeps the entry first, read from its source: it builds nothing, so a gate runs it."""
from pathlib import Path
import unittest

repository = Path(__file__).resolve().parents[4]


# A replay from an entry file rebuilds the program's files from the checker's resolved graph, whose order puts the
# entry's dependencies before it. Lowering reads the first file as the entry and routes every module identity from it,
# so load.go.txt's LatentReplayEntryReach keeps the entry first, as entry_overlay.py's LatentEntryReach does, and
# entry_test.py's test_overlay_keeps_the_entry_first reads it the same way. The order was observed on real products by
# an entry-root replay of a nested unit, red on the route panic under the old reach (J-vqjr7hz, J-t85bk9h; #4ft3dyf).
class EntryRootReachSourceTests(unittest.TestCase):
    def test_replay_reach_keeps_the_entry_first(self):
        text = (repository / 'stage3/census/latent/load.go.txt').read_text()
        reach = text[text.index('func (p *Program) LatentReplayEntryReach()'):]
        reach = reach[:reach.index('\n}\n') + 3]
        self.assertIn('p.files = []*ast.SourceFile{entry}', reach)
        self.assertIn('source != entry', reach)
        self.assertLess(reach.index('entry := p.files[0]'), reach.index('p.files = []*ast.SourceFile{entry}'))


if __name__ == '__main__':
    unittest.main()

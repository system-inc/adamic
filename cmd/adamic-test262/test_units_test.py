import unittest
import test_units


class TestUnitShards(unittest.TestCase):
    def test_every_unit_exactly_once(self):
        units = ["TestMiniRunner", "TestRegExpRunnerNode", "TestVerdictCorpus"]
        self.assertEqual(test_units.plan(units), [[unit] for unit in units])
        for mutant in (
            [[units[0]], [units[1]]],
            [[units[0]], [units[1]], [units[2]], [units[0]]],
            [[units[0]], [units[1]], [units[2]], ["TestUnknown"]],
        ):
            with self.subTest(mutant=mutant):
                with self.assertRaisesRegex(ValueError, "invalid coverage"):
                    test_units.coverage(units, mutant)
        with self.assertRaisesRegex(ValueError, "duplicate unit"):
            test_units.plan([units[0], units[0]])

    def test_completion_and_time_limit(self):
        events = [{"Action": "pass", "Test": name, "Elapsed": elapsed} for name, elapsed in
                  [("TestMiniRunner/pass", 1), ("TestMiniRunner", 3)]]
        self.assertTrue(test_units.observations("TestMiniRunner", events, 3.1)["valid"])
        self.assertTrue(test_units.observations("TestMiniRunner", events, 30)["valid"])
        self.assertFalse(test_units.observations("TestMiniRunner", events, 30.001)["valid"])
        too_slow = [dict(events[0], Elapsed=30.001), events[1]]
        self.assertFalse(test_units.observations("TestMiniRunner", too_slow, 3.1)["valid"])
        failed = [dict(events[0], Action="fail"), dict(events[1], Action="fail")]
        self.assertFalse(test_units.observations("TestMiniRunner", failed, 3.1)["valid"])
        with self.assertRaisesRegex(ValueError, "did not complete"):
            test_units.observations("TestMiniRunner", events[:1], 3.1)
        with self.assertRaisesRegex(ValueError, "duplicate completion"):
            test_units.observations("TestMiniRunner", events + events[1:], 3.1)
        with self.assertRaisesRegex(ValueError, "unexpected test"):
            test_units.observations("TestMiniRunner", [dict(events[0], Test="TestOther")], 3.1)


if __name__ == "__main__":
    unittest.main()

import unittest
import group

class GroupTest(unittest.TestCase):
    def test_counts_and_ownership(self):
        report = {"filters": [{"refusalReasons": [
            {"reason": "not yet: new String needs indexed exotic properties", "count": 42},
            {"reason": "not yet: a try around repeat, whose failure is a panic", "count": 2},
            {"reason": "refuses a method read as a value (match would lose its object)", "count": 3},
            {"reason": "not yet: String ToPrimitive needs a conversion member hidden by the object view", "count": 3},
        ]}]}
        result = group.grouped(report)
        self.assertEqual(sum(result["owners"].values()), 50)
        self.assertEqual(result["owners"]["compiler"], 42)
        self.assertEqual(result["owners"]["library"], 2)
        self.assertEqual(result["owners"]["objects/ruling"], 3)
        self.assertEqual(result["owners"]["regex/compiler method identity"], 3)

if __name__ == "__main__":
    unittest.main()

"""Prove overlay-off C and JavaScript identity against the given compiler base."""
from pathlib import Path
import runpy
runpy.run_path(str(Path(__file__).resolve().parents[1]/'speculative/isolation.py'), run_name='__main__')

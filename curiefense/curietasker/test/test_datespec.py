from datetime import datetime
from curietasker.task import time_match


def test_datespec():
    d = datetime(2020, 1, 1, 1, 1, 1)
    assert time_match(d, {"min": 1})

from pyproc_worker import health


def test_health_returns_fields():
    resp = health({})
    assert isinstance(resp, dict)
    assert resp.get("status") == "healthy"
    assert "pid" in resp

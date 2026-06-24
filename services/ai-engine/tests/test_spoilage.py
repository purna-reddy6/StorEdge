from app.spoilage.engine import SpoilageInput, assess_spoilage_risk


def make_input(**kwargs) -> SpoilageInput:
    defaults = dict(
        commodity_type="potato",
        temperature_celsius=4.0,
        relative_humidity=80.0,
        ethylene_ppm=0.0,
        storage_duration_days=30,
    )
    defaults.update(kwargs)
    return SpoilageInput(**defaults)


def test_ideal_conditions_safe():
    result = assess_spoilage_risk(make_input())
    assert result.risk_level == "safe"
    assert result.spoilage_risk_score < 0.15


def test_high_temperature_increases_risk():
    ideal = assess_spoilage_risk(make_input(temperature_celsius=4.0))
    hot = assess_spoilage_risk(make_input(temperature_celsius=20.0))
    assert hot.spoilage_risk_score > ideal.spoilage_risk_score


def test_high_ethylene_increases_risk_for_fruits():
    safe = assess_spoilage_risk(make_input(commodity_type="fruits", ethylene_ppm=0.1))
    at_risk = assess_spoilage_risk(make_input(commodity_type="fruits", ethylene_ppm=2.0))
    assert at_risk.spoilage_risk_score > safe.spoilage_risk_score


def test_score_clamped_between_0_and_1():
    result = assess_spoilage_risk(make_input(
        temperature_celsius=40.0, relative_humidity=100.0,
        ethylene_ppm=10.0, storage_duration_days=500,
    ))
    assert 0.0 <= result.spoilage_risk_score <= 1.0


def test_critical_risk_has_recommendations():
    result = assess_spoilage_risk(make_input(
        commodity_type="fruits",
        temperature_celsius=25.0,
        ethylene_ppm=5.0,
        storage_duration_days=25,
    ))
    assert result.risk_level in ("high", "critical")
    assert len(result.recommended_actions) > 0


def test_risk_levels_mapping():
    levels = {
        "safe": make_input(temperature_celsius=4.0, storage_duration_days=1),
    }
    for expected, inp in levels.items():
        result = assess_spoilage_risk(inp)
        assert result.risk_level == expected, f"Expected {expected}, got {result.risk_level}"

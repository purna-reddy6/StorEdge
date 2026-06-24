from app.pricing.engine import PricingInput, compute_dynamic_price, ALPHA, BETA, TARGET_UTILIZATION


def _price(utilization: float, volatility: float = 0.0, base: float = 2500.0) -> float:
    result = compute_dynamic_price(PricingInput(
        warehouse_id="w1",
        base_price_inr=base,
        occupancy_rate=utilization,
        commodity_volatility=volatility,
    ))
    return result.dynamic_price_inr


def test_price_at_target_utilization_equals_base():
    p = _price(TARGET_UTILIZATION)
    assert abs(p - 2500.0) < 1.0, f"Expected ~2500, got {p}"


def test_price_increases_above_target():
    p_low = _price(0.70)
    p_high = _price(0.95)
    assert p_high > p_low, "Higher utilization should produce higher price"


def test_price_clamped_at_minimum():
    p = _price(0.0)
    assert p >= 2500 * 0.70, "Price must not drop below 70% of base"


def test_price_clamped_at_maximum():
    p = _price(1.0, volatility=1.0)
    assert p <= 2500 * 2.50, "Price must not exceed 250% of base"


def test_volatility_increases_price():
    p_no_vol = _price(0.90, volatility=0.0)
    p_with_vol = _price(0.90, volatility=0.5)
    assert p_with_vol > p_no_vol, "Volatility should push price up"


def test_pricing_formula():
    """Exact formula verification: P = base * clamp(1 + α*(U-U*) + β*V, 0.70, 2.50)"""
    U = 0.90
    V = 0.0
    base = 2000.0
    expected_multiplier = max(0.70, min(2.50, 1.0 + ALPHA * (U - TARGET_UTILIZATION) + BETA * V))
    expected = round(base * expected_multiplier, 2)
    actual = _price(U, V, base)
    assert abs(actual - expected) < 1.0, f"Expected {expected}, got {actual}"

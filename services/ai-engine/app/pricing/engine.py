"""
Dynamic Pricing Engine
Blueprint formula (Part 9, Predictive Dynamic Pricing Algorithm):
    P = P_base × (1 + α(U - U*) + β×V)

Where:
    P_base = warehouse base rental rate
    U      = current occupancy rate (0.0 - 1.0)
    U*     = optimal target utilization (default: 0.85)
    V      = spot market price volatility of stored commodity
    α, β   = scaling parameters (trained via RL; hardcoded for MVP)
"""

from dataclasses import dataclass
from typing import Optional


TARGET_UTILIZATION = 0.85  # U* — optimal occupancy from blueprint
ALPHA = 0.40               # Occupancy elasticity coefficient
BETA = 0.15                # Commodity volatility coefficient
MIN_PRICE_MULTIPLIER = 0.70  # Never price below 70% of base (floor)
MAX_PRICE_MULTIPLIER = 2.50  # Never price above 250% of base (ceiling)


@dataclass
class PricingInput:
    warehouse_id: str
    base_price_inr: float
    occupancy_rate: float             # 0.0 to 1.0 (e.g., 0.78 = 78% full)
    commodity_volatility: float = 0.0 # 0.0 to 1.0 — sourced from commodity spot markets
    season_multiplier: float = 1.0    # 1.0 = normal; 1.3 = harvest peak; 0.85 = off-peak


@dataclass
class PricingOutput:
    warehouse_id: str
    dynamic_price_inr: float
    base_price_inr: float
    occupancy_rate: float
    price_multiplier: float
    pricing_tier: str       # "peak" | "standard" | "off-peak"
    price_delta_inr: float  # Change from base


def compute_dynamic_price(inp: PricingInput) -> PricingOutput:
    """
    Applies the blueprint pricing formula with safety guardrails.
    Guardrail: prevents erratic spot-pricing loops (blueprint risk note, Part 7).
    """
    occupancy_deviation = inp.occupancy_rate - TARGET_UTILIZATION
    volatility_adjustment = BETA * inp.commodity_volatility

    raw_multiplier = 1.0 + ALPHA * occupancy_deviation + volatility_adjustment
    raw_multiplier *= inp.season_multiplier

    # Clamp to safe range to prevent erratic pricing
    multiplier = max(MIN_PRICE_MULTIPLIER, min(MAX_PRICE_MULTIPLIER, raw_multiplier))
    dynamic_price = round(inp.base_price_inr * multiplier, 2)

    tier = _pricing_tier(inp.occupancy_rate)

    return PricingOutput(
        warehouse_id=inp.warehouse_id,
        dynamic_price_inr=dynamic_price,
        base_price_inr=inp.base_price_inr,
        occupancy_rate=inp.occupancy_rate,
        price_multiplier=round(multiplier, 4),
        pricing_tier=tier,
        price_delta_inr=round(dynamic_price - inp.base_price_inr, 2),
    )


def _pricing_tier(occupancy: float) -> str:
    if occupancy >= 0.90:
        return "peak"
    elif occupancy < 0.60:
        return "off-peak"
    return "standard"


def estimate_monthly_cost(price_per_pallet: float, pallet_count: int, duration_days: int) -> float:
    """Pro-rated monthly cost calculation for booking estimator."""
    months = duration_days / 30.0
    return round(price_per_pallet * pallet_count * months, 2)

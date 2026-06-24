"""
Crop Spoilage & Quality Classifier
Blueprint: 3D-CNN + LSTM model (Part 7).
Phase 1 MVP: rule-based risk scorer using sensor telemetry.
Phase 2: Full 3D-CNN/LSTM trained on hyperspectral images + telemetry.

Risk factors:
- Temperature breach (above max or below min for commodity)
- High ethylene concentration (early decay signal)
- High humidity (mold risk)
- Duration exceeding expected shelf life
"""

from dataclasses import dataclass
from typing import Optional
from datetime import datetime, timedelta


COMMODITY_PROFILES = {
    "potato": {
        "optimal_temp_min": 2.0, "optimal_temp_max": 6.0,
        "max_humidity": 90.0, "ethylene_sensitivity": "low",
        "shelf_life_days": 180,
    },
    "onion": {
        "optimal_temp_min": 0.0, "optimal_temp_max": 5.0,
        "max_humidity": 75.0, "ethylene_sensitivity": "low",
        "shelf_life_days": 120,
    },
    "fruits": {
        "optimal_temp_min": 0.0, "optimal_temp_max": 8.0,
        "max_humidity": 90.0, "ethylene_sensitivity": "high",
        "shelf_life_days": 30,
    },
    "vegetables": {
        "optimal_temp_min": 1.0, "optimal_temp_max": 7.0,
        "max_humidity": 95.0, "ethylene_sensitivity": "medium",
        "shelf_life_days": 21,
    },
    "pharma": {
        "optimal_temp_min": 2.0, "optimal_temp_max": 8.0,
        "max_humidity": 60.0, "ethylene_sensitivity": "none",
        "shelf_life_days": 365,
    },
}


@dataclass
class SpoilageInput:
    commodity_type: str
    temperature_celsius: float
    relative_humidity: float
    ethylene_ppm: float
    storage_duration_days: int
    inward_date: Optional[datetime] = None


@dataclass
class SpoilageAssessment:
    commodity_type: str
    spoilage_risk_score: float      # 0.0 (safe) to 1.0 (critical)
    risk_level: str                  # "safe" | "low" | "medium" | "high" | "critical"
    risk_factors: list
    estimated_shelf_life_remaining: int  # days
    recommended_actions: list


def assess_spoilage_risk(inp: SpoilageInput) -> SpoilageAssessment:
    """
    Compute spoilage risk score from sensor telemetry.
    Blueprint target: reduce agricultural spoilage losses by >20% (Part 7).
    """
    profile = COMMODITY_PROFILES.get(inp.commodity_type.lower(), COMMODITY_PROFILES["vegetables"])
    risk_factors = []
    risk_score = 0.0

    # Temperature deviation
    if inp.temperature_celsius > profile["optimal_temp_max"]:
        excess = inp.temperature_celsius - profile["optimal_temp_max"]
        temp_risk = min(0.4, excess / 10.0)
        risk_score += temp_risk
        risk_factors.append({
            "factor": "temperature_high",
            "severity": "high" if excess > 3 else "medium",
            "detail": f"Temperature {inp.temperature_celsius}°C exceeds optimal max {profile['optimal_temp_max']}°C",
        })
    elif inp.temperature_celsius < profile["optimal_temp_min"]:
        deficit = profile["optimal_temp_min"] - inp.temperature_celsius
        temp_risk = min(0.25, deficit / 10.0)
        risk_score += temp_risk
        risk_factors.append({
            "factor": "temperature_low",
            "severity": "medium",
            "detail": f"Temperature {inp.temperature_celsius}°C below optimal min {profile['optimal_temp_min']}°C",
        })

    # Ethylene concentration
    ethylene_thresholds = {"high": 0.5, "medium": 0.8, "low": 1.5, "none": 999}
    eth_threshold = ethylene_thresholds.get(profile["ethylene_sensitivity"], 1.0)
    if inp.ethylene_ppm > eth_threshold:
        eth_risk = min(0.35, (inp.ethylene_ppm / eth_threshold) * 0.15)
        risk_score += eth_risk
        risk_factors.append({
            "factor": "ethylene_elevated",
            "severity": "critical" if inp.ethylene_ppm > eth_threshold * 2 else "high",
            "detail": f"Ethylene {inp.ethylene_ppm:.2f} ppm (threshold: {eth_threshold} ppm) — early decay",
        })

    # Humidity
    if inp.relative_humidity > profile["max_humidity"]:
        hum_excess = inp.relative_humidity - profile["max_humidity"]
        hum_risk = min(0.20, hum_excess / 50.0)
        risk_score += hum_risk
        risk_factors.append({
            "factor": "humidity_high",
            "severity": "medium",
            "detail": f"Humidity {inp.relative_humidity}% exceeds max {profile['max_humidity']}% — mold risk",
        })

    # Time elapsed vs. shelf life
    shelf_life = profile["shelf_life_days"]
    if inp.storage_duration_days > 0:
        time_fraction = inp.storage_duration_days / shelf_life
        time_risk = min(0.30, time_fraction * 0.25)
        risk_score += time_risk
        remaining = max(0, shelf_life - inp.storage_duration_days)
        if remaining < 14:
            risk_factors.append({
                "factor": "approaching_shelf_life",
                "severity": "high" if remaining < 7 else "medium",
                "detail": f"Only {remaining} days of shelf life remaining",
            })

    risk_score = min(1.0, risk_score)
    remaining_days = max(0, profile["shelf_life_days"] - inp.storage_duration_days)
    remaining_days = int(remaining_days * (1.0 - risk_score * 0.5))

    return SpoilageAssessment(
        commodity_type=inp.commodity_type,
        spoilage_risk_score=round(risk_score, 3),
        risk_level=_risk_level(risk_score),
        risk_factors=risk_factors,
        estimated_shelf_life_remaining=remaining_days,
        recommended_actions=_get_recommendations(risk_score, risk_factors),
    )


def _risk_level(score: float) -> str:
    if score < 0.15:
        return "safe"
    elif score < 0.30:
        return "low"
    elif score < 0.50:
        return "medium"
    elif score < 0.70:
        return "high"
    return "critical"


def _get_recommendations(score: float, factors: list) -> list:
    actions = []
    factor_types = {f["factor"] for f in factors}

    if "temperature_high" in factor_types:
        actions.append("Reduce cooling set-point by 1-2°C immediately")
        actions.append("Check refrigeration unit performance and coolant levels")
    if "ethylene_elevated" in factor_types:
        actions.append("Activate ethylene scrubbers or increase air exchange rate")
        actions.append("Segregate highly ethylene-sensitive produce")
    if "humidity_high" in factor_types:
        actions.append("Enable dehumidification units")
        actions.append("Check for water ingress or condensation sources")
    if "approaching_shelf_life" in factor_types:
        actions.append("Prioritize immediate dispatch or sale of stored goods")
        actions.append("Notify farmer/trader to plan outward within 7 days")
    if score >= 0.7:
        actions.append("URGENT: Alert warehouse operator and tenant immediately")
        actions.append("Consider emergency partial sale through commodity trading platform")

    return actions

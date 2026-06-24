"""
Smart Spatial Recommendation Engine
Blueprint: LightGBM Ranker + Multi-Criteria Decision-Making (MCDM) neural network.
Phase 1 MVP: rule-based MCDM scoring mirroring the Go scorer's formula.
Phase 2: LightGBM model trained on actual booking conversion data.
"""

from dataclasses import dataclass, field
from typing import List, Optional
import math


@dataclass
class WarehouseCandidate:
    warehouse_id: str
    distance_km: float
    current_price_inr: float
    rating: float                     # 0.0 - 5.0
    review_count: int
    wdra_registered: bool
    apmc_licensed: bool
    available_pallets: int
    total_pallets: int
    warehouse_type: str
    min_temperature: Optional[float] = None
    max_temperature: Optional[float] = None


@dataclass
class TenantPreferences:
    required_pallets: int
    max_distance_km: float = 50.0
    max_price_inr: float = 0.0
    wdra_required: bool = False
    needs_cold_chain: bool = False
    min_temperature: Optional[float] = None
    max_temperature: Optional[float] = None

    # Blueprint weights: w1=distance, w2=price, w3=quality, w4=temperature
    weight_distance: float = 0.25
    weight_price: float = 0.25
    weight_quality: float = 0.25
    weight_temperature: float = 0.25


@dataclass
class ScoredWarehouse:
    warehouse_id: str
    match_score: float          # S ∈ [0, 1]
    distance_score: float
    price_score: float
    quality_score: float
    temperature_score: float
    rank: int = 0


def score_warehouses(
    candidates: List[WarehouseCandidate],
    prefs: TenantPreferences,
) -> List[ScoredWarehouse]:
    """
    Score and rank warehouses using the blueprint MCDM formula:
    S = w1×D_norm + w2×P_norm + w3×Q_norm + w4×T_compat
    """
    # Normalize weights
    total = prefs.weight_distance + prefs.weight_price + prefs.weight_quality + prefs.weight_temperature
    if total == 0:
        w1, w2, w3, w4 = 0.25, 0.25, 0.25, 0.25
    else:
        w1 = prefs.weight_distance / total
        w2 = prefs.weight_price / total
        w3 = prefs.weight_quality / total
        w4 = prefs.weight_temperature / total

    max_price = prefs.max_price_inr or 5000.0
    max_dist = prefs.max_distance_km or 100.0

    scored = []
    for c in candidates:
        if c.available_pallets < prefs.required_pallets:
            continue  # Filter: insufficient capacity

        d_norm = _score_distance(c.distance_km, max_dist)
        p_norm = _score_price(c.current_price_inr, max_price)
        q_norm = _score_quality(c.rating, c.review_count, c.wdra_registered, c.apmc_licensed)
        t_compat = _score_temperature(c, prefs)

        total_score = w1 * d_norm + w2 * p_norm + w3 * q_norm + w4 * t_compat

        scored.append(ScoredWarehouse(
            warehouse_id=c.warehouse_id,
            match_score=round(total_score, 4),
            distance_score=round(d_norm, 4),
            price_score=round(p_norm, 4),
            quality_score=round(q_norm, 4),
            temperature_score=round(t_compat, 4),
        ))

    scored.sort(key=lambda x: x.match_score, reverse=True)
    for i, s in enumerate(scored):
        s.rank = i + 1

    return scored


def _score_distance(dist_km: float, max_km: float) -> float:
    """Closer = higher score. 0.0 at max_km, 1.0 at 0km."""
    return max(0.0, min(1.0, 1.0 - (dist_km / max_km)))


def _score_price(price: float, max_price: float) -> float:
    """Cheaper = higher score."""
    return max(0.0, min(1.0, 1.0 - (price / max_price)))


def _score_quality(rating: float, reviews: int, wdra: bool, apmc: bool) -> float:
    """Combines rating credibility + compliance bonuses."""
    rating_norm = rating / 5.0
    credibility = 1.0 - math.exp(-reviews / 30.0)
    adjusted = rating_norm * credibility

    compliance = 0.12 if wdra else 0.0
    compliance += 0.05 if apmc else 0.0

    return min(1.0, adjusted + compliance)


def _score_temperature(c: WarehouseCandidate, prefs: TenantPreferences) -> float:
    """Returns temperature compatibility score."""
    if not prefs.needs_cold_chain:
        return 0.5  # Neutral — temperature not a factor

    has_cold = c.min_temperature is not None and c.max_temperature is not None
    if not has_cold:
        return 0.0  # No cold capability

    if prefs.min_temperature is not None and prefs.max_temperature is not None:
        if c.min_temperature <= prefs.min_temperature and c.max_temperature >= prefs.max_temperature:
            return 1.0  # Perfect match
        return 0.3  # Partial overlap

    return 0.7  # Has cold chain but no precise range required

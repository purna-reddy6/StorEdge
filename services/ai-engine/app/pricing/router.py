from fastapi import APIRouter
from pydantic import BaseModel, Field
from typing import Optional

from app.pricing.engine import PricingInput, compute_dynamic_price, estimate_monthly_cost

router = APIRouter()


class ComputePriceRequest(BaseModel):
    warehouse_id: str
    base_price_inr: float = Field(..., gt=0)
    occupancy_rate: float = Field(..., ge=0.0, le=1.0)
    commodity_volatility: float = Field(0.0, ge=0.0, le=1.0)
    season_multiplier: float = Field(1.0, gt=0)


class ComputePriceResponse(BaseModel):
    warehouse_id: str
    dynamic_price_inr: float
    base_price_inr: float
    occupancy_rate: float
    price_multiplier: float
    pricing_tier: str
    price_delta_inr: float


@router.post("/compute", response_model=ComputePriceResponse)
def compute_price(req: ComputePriceRequest):
    """
    Compute dynamic pallet rental price using the blueprint formula:
    P = P_base × (1 + α(U - U*) + β×V)
    Called by the Go search-match service every 15 minutes per warehouse.
    """
    inp = PricingInput(
        warehouse_id=req.warehouse_id,
        base_price_inr=req.base_price_inr,
        occupancy_rate=req.occupancy_rate,
        commodity_volatility=req.commodity_volatility,
        season_multiplier=req.season_multiplier,
    )
    result = compute_dynamic_price(inp)
    return result


class CostEstimateRequest(BaseModel):
    price_per_pallet_inr: float
    pallet_count: int = Field(..., ge=1)
    duration_days: int = Field(..., ge=1)


@router.post("/estimate-cost")
def estimate_cost(req: CostEstimateRequest):
    """Pro-rated monthly cost estimator shown to tenants before booking."""
    cost = estimate_monthly_cost(req.price_per_pallet_inr, req.pallet_count, req.duration_days)
    return {
        "estimated_total_inr": cost,
        "per_pallet_inr": req.price_per_pallet_inr,
        "pallet_count": req.pallet_count,
        "duration_days": req.duration_days,
    }

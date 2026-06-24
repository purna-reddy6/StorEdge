from fastapi import APIRouter
from pydantic import BaseModel, Field
from typing import List, Optional

from app.recommender.engine import (
    WarehouseCandidate, TenantPreferences, ScoredWarehouse, score_warehouses
)

router = APIRouter()


class WarehouseCandidateInput(BaseModel):
    warehouse_id: str
    distance_km: float
    current_price_inr: float
    rating: float = Field(0.0, ge=0.0, le=5.0)
    review_count: int = 0
    wdra_registered: bool = False
    apmc_licensed: bool = False
    available_pallets: int
    total_pallets: int
    warehouse_type: str
    min_temperature: Optional[float] = None
    max_temperature: Optional[float] = None


class TenantPreferencesInput(BaseModel):
    required_pallets: int = Field(..., ge=1)
    max_distance_km: float = 50.0
    max_price_inr: float = 0.0
    wdra_required: bool = False
    needs_cold_chain: bool = False
    min_temperature: Optional[float] = None
    max_temperature: Optional[float] = None
    weight_distance: float = Field(0.25, ge=0.0, le=1.0)
    weight_price: float = Field(0.25, ge=0.0, le=1.0)
    weight_quality: float = Field(0.25, ge=0.0, le=1.0)
    weight_temperature: float = Field(0.25, ge=0.0, le=1.0)


class RankRequest(BaseModel):
    candidates: List[WarehouseCandidateInput]
    preferences: TenantPreferencesInput


class RankResponse(BaseModel):
    ranked_warehouses: List[dict]
    total_candidates: int
    total_eligible: int


@router.post("/rank", response_model=RankResponse)
def rank_warehouses(req: RankRequest):
    """
    Rank warehouse candidates using the MCDM scoring formula.
    Called by search-match service to apply AI-layer ranking on top of PostGIS results.
    """
    candidates = [
        WarehouseCandidate(
            warehouse_id=c.warehouse_id,
            distance_km=c.distance_km,
            current_price_inr=c.current_price_inr,
            rating=c.rating,
            review_count=c.review_count,
            wdra_registered=c.wdra_registered,
            apmc_licensed=c.apmc_licensed,
            available_pallets=c.available_pallets,
            total_pallets=c.total_pallets,
            warehouse_type=c.warehouse_type,
            min_temperature=c.min_temperature,
            max_temperature=c.max_temperature,
        )
        for c in req.candidates
    ]

    prefs = TenantPreferences(
        required_pallets=req.preferences.required_pallets,
        max_distance_km=req.preferences.max_distance_km,
        max_price_inr=req.preferences.max_price_inr,
        wdra_required=req.preferences.wdra_required,
        needs_cold_chain=req.preferences.needs_cold_chain,
        min_temperature=req.preferences.min_temperature,
        max_temperature=req.preferences.max_temperature,
        weight_distance=req.preferences.weight_distance,
        weight_price=req.preferences.weight_price,
        weight_quality=req.preferences.weight_quality,
        weight_temperature=req.preferences.weight_temperature,
    )

    scored = score_warehouses(candidates, prefs)

    return RankResponse(
        ranked_warehouses=[
            {
                "rank": s.rank,
                "warehouse_id": s.warehouse_id,
                "match_score": s.match_score,
                "score_breakdown": {
                    "distance": s.distance_score,
                    "price": s.price_score,
                    "quality": s.quality_score,
                    "temperature": s.temperature_score,
                },
            }
            for s in scored
        ],
        total_candidates=len(candidates),
        total_eligible=len(scored),
    )

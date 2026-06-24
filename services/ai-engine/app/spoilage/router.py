from fastapi import APIRouter
from pydantic import BaseModel, Field
from typing import Optional
from datetime import datetime

from app.spoilage.engine import SpoilageInput, assess_spoilage_risk

router = APIRouter()


class SpoilageRequest(BaseModel):
    commodity_type: str
    temperature_celsius: float
    relative_humidity: float = Field(..., ge=0.0, le=100.0)
    ethylene_ppm: float = Field(0.0, ge=0.0)
    storage_duration_days: int = Field(0, ge=0)
    inward_date: Optional[datetime] = None


@router.post("/assess")
def assess_spoilage(req: SpoilageRequest):
    """
    Assess crop spoilage risk from real-time sensor telemetry.
    Called by IoT gateway when threshold alerts are triggered.
    Blueprint target: reduce agricultural spoilage by >20%.
    """
    inp = SpoilageInput(
        commodity_type=req.commodity_type,
        temperature_celsius=req.temperature_celsius,
        relative_humidity=req.relative_humidity,
        ethylene_ppm=req.ethylene_ppm,
        storage_duration_days=req.storage_duration_days,
        inward_date=req.inward_date,
    )
    result = assess_spoilage_risk(inp)
    return {
        "commodity_type": result.commodity_type,
        "spoilage_risk_score": result.spoilage_risk_score,
        "risk_level": result.risk_level,
        "risk_factors": result.risk_factors,
        "estimated_shelf_life_remaining_days": result.estimated_shelf_life_remaining,
        "recommended_actions": result.recommended_actions,
    }

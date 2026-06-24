"""
StorEdge AI Engine — FastAPI service
Implements blueprint Part 7 AI features:
- Dynamic pricing engine (P = P_base × (1 + α(U - U*) + β×V))
- Smart spatial recommendation (LightGBM + MCDM)
- Spoilage risk scorer
"""

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
import uvicorn
import os
from dotenv import load_dotenv

from app.pricing.router import router as pricing_router
from app.recommender.router import router as recommender_router
from app.spoilage.router import router as spoilage_router

load_dotenv()

app = FastAPI(
    title="StorEdge AI Engine",
    description="Dynamic pricing, warehouse recommendation, and crop spoilage prediction",
    version="0.1.0",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(pricing_router, prefix="/api/v1/pricing", tags=["Pricing"])
app.include_router(recommender_router, prefix="/api/v1/recommender", tags=["Recommender"])
app.include_router(spoilage_router, prefix="/api/v1/spoilage", tags=["Spoilage"])


@app.get("/health")
def health():
    return {"status": "ok", "service": "ai-engine"}


if __name__ == "__main__":
    port = int(os.getenv("PORT", 8084))
    uvicorn.run("app.main:app", host="0.0.0.0", port=port, reload=True)

package matching

import "math"

// Score computes the multi-dimensional warehouse suitability score from the blueprint:
//
//	S = w1×D_norm + w2×P_norm + w3×Q_norm + w4×T_compat
//
// Where each factor is normalized to [0, 1] and weights w1..w4 are tenant-defined.
// Higher score = better match.
func Score(w Warehouse, req SearchRequest) float64 {
	w1 := normalizeWeights(req.WeightDistance, req.WeightPrice, req.WeightQuality, req.WeightTemperature)

	dNorm := scoreDistance(w.DistanceKm, req.MaxDistanceKm)
	pNorm := scorePrice(w.CurrentPrice, req.MaxPricePerPallet)
	qNorm := scoreQuality(w.Rating, w.TotalReviews, w.GSTRegistered)
	tCompat := scoreTemperature(w, req)

	return w1[0]*dNorm + w1[1]*pNorm + w1[2]*qNorm + w1[3]*tCompat
}

// normalizeWeights ensures the four weights sum to 1.0.
// Falls back to equal weights if all are zero or unset.
func normalizeWeights(d, p, q, t float64) [4]float64 {
	total := d + p + q + t
	if total == 0 {
		return [4]float64{0.25, 0.25, 0.25, 0.25}
	}
	return [4]float64{d / total, p / total, q / total, t / total}
}

// scoreDistance: closer is better. Returns 1.0 at distance=0, 0.0 at maxDistance.
func scoreDistance(distKm, maxKm float64) float64 {
	if maxKm == 0 {
		maxKm = 100
	}
	normalized := 1.0 - (distKm / maxKm)
	return math.Max(0, math.Min(1, normalized))
}

// scorePrice: cheaper is better relative to max acceptable price.
// Returns 1.0 at lowest possible price, 0.0 at maxPrice.
func scorePrice(currentPrice, maxPrice float64) float64 {
	if maxPrice == 0 {
		// No price constraint — normalize against a 5000 INR/pallet ceiling
		maxPrice = 5000
	}
	normalized := 1.0 - (currentPrice / maxPrice)
	return math.Max(0, math.Min(1, normalized))
}

// scoreQuality combines rating (0-5), review count (credibility), and GST registration.
func scoreQuality(rating float64, reviews int, gstRegistered bool) float64 {
	ratingScore := rating / 5.0

	// Review credibility (logarithmic scaling: 50 reviews = 0.85 credibility)
	credibility := 1.0 - math.Exp(-float64(reviews)/30.0)
	adjustedRating := ratingScore * credibility

	// GST registration signals a verified, tax-compliant host
	compliance := 0.0
	if gstRegistered {
		compliance += 0.12
	}

	return math.Min(1.0, adjustedRating+compliance)
}

// scoreTemperature: 1.0 if the warehouse temperature range is compatible,
// 0.0 if there's no cold-chain requirement (neutral) or incompatible range.
func scoreTemperature(w Warehouse, req SearchRequest) float64 {
	needsColdChain := req.MinTemperature != 0 || req.MaxTemperature != 0
	if !needsColdChain {
		return 0.5 // neutral — temperature not a factor for this tenant
	}

	hasColdCapability := w.MinTemperature != -99 && w.MaxTemperature != 99

	if !hasColdCapability {
		return 0.0 // tenant needs cold chain, warehouse has none
	}

	// Check if warehouse range covers the required range
	if w.MinTemperature <= req.MinTemperature && w.MaxTemperature >= req.MaxTemperature {
		return 1.0 // perfect match
	}

	return 0.3 // partial overlap (may still work with caveats)
}

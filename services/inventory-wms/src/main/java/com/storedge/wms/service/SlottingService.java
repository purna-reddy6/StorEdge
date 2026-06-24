package com.storedge.wms.service;

import com.storedge.wms.repository.PalletItemRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;
import java.util.UUID;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.ConcurrentHashMap;

/**
 * SlottingService assigns optimal warehouse slot positions for incoming pallets.
 *
 * Implements the contextual multi-armed bandit (LinUCB) approach from the blueprint
 * (Part 7, AI Features) for Phase 2. Phase 1 uses a simpler zone-based heuristic:
 * - High-velocity commodities (pharma, electronics) → Zone A (near loading bay)
 * - Agricultural produce → Zone B (cold chain proximity)
 * - Bulk/long-term → Zone C (rear storage)
 */
@Service
@RequiredArgsConstructor
@Slf4j
public class SlottingService {

    private final PalletItemRepository palletRepo;

    // Zone counters — in Phase 2, replace with LinUCB bandit model
    private final ConcurrentHashMap<String, AtomicInteger> zoneCounters = new ConcurrentHashMap<>();

    /**
     * Assigns an optimal slot position based on commodity type and weight.
     * Format: "ZONE-ROW-LEVEL" e.g., "A-03-02"
     */
    public String assignOptimalSlot(UUID warehouseId, String commodityType, BigDecimal weightKg) {
        String zone = determineZone(commodityType);
        String key = warehouseId.toString() + ":" + zone;

        AtomicInteger counter = zoneCounters.computeIfAbsent(key, k -> new AtomicInteger(1));
        int position = counter.getAndIncrement();

        int row = ((position - 1) / 5) + 1;   // 5 slots per row
        int level = ((position - 1) % 5) + 1;

        String slot = String.format("%s-%02d-%02d", zone, row, level);
        log.debug("Assigned slot {} to commodity {} in warehouse {}", slot, commodityType, warehouseId);
        return slot;
    }

    private String determineZone(String commodityType) {
        return switch (commodityType.toLowerCase()) {
            // High-velocity → Zone A (closest to loading bay)
            case "pharma", "electronics", "fmcg" -> "A";
            // Agricultural produce → Zone B (cold storage proximity)
            case "potato", "onion", "garlic", "fruits", "vegetables", "dairy" -> "B";
            // Bulk / long-term → Zone C
            default -> "C";
        };
    }
}

package com.storedge.wms.controller;

import com.storedge.wms.model.WarehouseFloorPlan;
import com.storedge.wms.repository.FloorPlanRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.time.OffsetDateTime;
import java.util.Optional;
import java.util.UUID;

@RestController
@RequestMapping("/api/v1/floor-plans")
@RequiredArgsConstructor
public class FloorPlanController {

    private final FloorPlanRepository floorPlanRepo;

    /** Get the interactive floor plan for a warehouse. */
    @GetMapping("/{warehouseId}")
    public ResponseEntity<WarehouseFloorPlan> getFloorPlan(@PathVariable UUID warehouseId) {
        Optional<WarehouseFloorPlan> plan = floorPlanRepo.findByWarehouseId(warehouseId);
        return plan.map(ResponseEntity::ok)
                   .orElse(ResponseEntity.notFound().build());
    }

    /** Create or update the floor plan grid (drag-and-drop from operator portal). */
    @PutMapping("/{warehouseId}")
    public ResponseEntity<WarehouseFloorPlan> upsertFloorPlan(
        @PathVariable UUID warehouseId,
        @RequestBody WarehouseFloorPlan plan
    ) {
        plan.setWarehouseId(warehouseId);
        plan.setUpdatedAt(OffsetDateTime.now());

        Optional<WarehouseFloorPlan> existing = floorPlanRepo.findByWarehouseId(warehouseId);
        if (existing.isPresent()) {
            plan.setId(existing.get().getId());
        }

        WarehouseFloorPlan saved = floorPlanRepo.save(plan);
        return ResponseEntity.status(existing.isPresent() ? HttpStatus.OK : HttpStatus.CREATED).body(saved);
    }
}

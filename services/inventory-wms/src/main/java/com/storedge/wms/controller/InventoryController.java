package com.storedge.wms.controller;

import com.storedge.wms.dto.InwardRequest;
import com.storedge.wms.dto.ReleaseRequestDto;
import com.storedge.wms.model.PalletItem;
import com.storedge.wms.model.StockReleaseRequest;
import com.storedge.wms.service.InventoryService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;
import java.util.UUID;

@RestController
@RequestMapping("/api/v1")
@RequiredArgsConstructor
@Slf4j
public class InventoryController {

    private final InventoryService inventoryService;

    // ─── Inward / Outward ──────────────────────────────────────────────────

    /** Record goods arriving at a warehouse (inward). */
    @PostMapping("/inventory/inward")
    public ResponseEntity<PalletItem> recordInward(@Valid @RequestBody InwardRequest req) {
        PalletItem item = inventoryService.recordInward(req);
        return ResponseEntity.status(HttpStatus.CREATED).body(item);
    }

    /** List all active inventory in a warehouse. */
    @GetMapping("/inventory/warehouse/{warehouseId}")
    public ResponseEntity<List<PalletItem>> getWarehouseInventory(@PathVariable UUID warehouseId) {
        return ResponseEntity.ok(inventoryService.getInventoryByWarehouse(warehouseId));
    }

    /** List all inventory for a tenant (farmer/trader). */
    @GetMapping("/inventory/tenant/{tenantId}")
    public ResponseEntity<List<PalletItem>> getTenantInventory(@PathVariable UUID tenantId) {
        return ResponseEntity.ok(inventoryService.getInventoryByTenant(tenantId));
    }

    /** Get active inventory count for a warehouse (for occupancy calculations). */
    @GetMapping("/inventory/warehouse/{warehouseId}/count")
    public ResponseEntity<Map<String, Long>> getInventoryCount(@PathVariable UUID warehouseId) {
        long count = inventoryService.getActiveInventoryCount(warehouseId);
        return ResponseEntity.ok(Map.of("active_pallet_count", count, "warehouse_id", (long) warehouseId.hashCode()));
    }

    // ─── OTP Stock Release ─────────────────────────────────────────────────

    /** Step 1: Farmer requests stock release — triggers OTP to their phone. */
    @PostMapping("/inventory/release/request")
    public ResponseEntity<StockReleaseRequest> requestRelease(@Valid @RequestBody ReleaseRequestDto req) {
        StockReleaseRequest release = inventoryService.initiateRelease(req);
        return ResponseEntity.status(HttpStatus.CREATED).body(release);
    }

    /** Step 2: Farmer submits OTP received on phone — authorizes the release remotely. */
    @PostMapping("/inventory/release/{releaseId}/authorize")
    public ResponseEntity<StockReleaseRequest> authorizeRelease(
        @PathVariable UUID releaseId,
        @RequestParam String otp,
        @RequestParam(required = false) UUID operatorId
    ) {
        StockReleaseRequest release = inventoryService.authorizeReleaseWithOTP(releaseId, otp, operatorId);
        return ResponseEntity.ok(release);
    }

    /** Step 3: Warehouse operator confirms physical goods have been removed. */
    @PostMapping("/inventory/release/{releaseId}/complete")
    public ResponseEntity<StockReleaseRequest> completeRelease(@PathVariable UUID releaseId) {
        StockReleaseRequest release = inventoryService.completeRelease(releaseId);
        return ResponseEntity.ok(release);
    }
}

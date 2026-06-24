package com.storedge.wms.service;

import com.storedge.wms.dto.InwardRequest;
import com.storedge.wms.dto.ReleaseRequestDto;
import com.storedge.wms.model.PalletItem;
import com.storedge.wms.model.StockReleaseRequest;
import com.storedge.wms.repository.PalletItemRepository;
import com.storedge.wms.repository.StockReleaseRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.OffsetDateTime;
import java.util.List;
import java.util.UUID;

@Service
@RequiredArgsConstructor
@Slf4j
public class InventoryService {

    private final PalletItemRepository palletRepo;
    private final StockReleaseRepository releaseRepo;
    private final SlottingService slottingService;
    private final OtpService otpService;
    private final KafkaEventService kafkaEventService;

    // ─── Inward ────────────────────────────────────────────────────────────

    @Transactional
    public PalletItem recordInward(InwardRequest req) {
        String slotPosition = req.getSlotPosition();
        if (slotPosition == null || slotPosition.isBlank()) {
            slotPosition = slottingService.assignOptimalSlot(
                req.getWarehouseId(), req.getCommodityType(), req.getWeightKg()
            );
        }

        PalletItem item = PalletItem.builder()
            .bookingId(req.getBookingId())
            .warehouseId(req.getWarehouseId())
            .tenantId(req.getTenantId())
            .commodityType(PalletItem.CommodityType.valueOf(req.getCommodityType()))
            .commodityDescription(req.getCommodityDescription())
            .weightKg(req.getWeightKg())
            .volumeCubicMeters(req.getVolumeCubicMeters())
            .bagCount(req.getBagCount())
            .slotPosition(slotPosition)
            .rfidTagId(req.getRfidTagId())
            .expectedOutwardDate(req.getExpectedOutwardDate())
            .inwardDate(OffsetDateTime.now())
            .build();

        PalletItem saved = palletRepo.save(item);

        kafkaEventService.publishInventoryEvent("PALLET_INWARD", saved);
        log.info("Pallet inward recorded: id={}, slot={}, warehouse={}",
            saved.getId(), slotPosition, req.getWarehouseId());

        return saved;
    }

    // ─── OTP Stock Release ─────────────────────────────────────────────────

    @Transactional
    public StockReleaseRequest initiateRelease(ReleaseRequestDto req) {
        PalletItem pallet = palletRepo.findById(req.getPalletItemId())
            .orElseThrow(() -> new IllegalArgumentException("Pallet not found: " + req.getPalletItemId()));

        if (!pallet.getTenantId().equals(req.getTenantId())) {
            throw new SecurityException("Tenant does not own this pallet");
        }
        if (pallet.isEnwrsPledged()) {
            throw new IllegalStateException("Pallet is pledged as e-NWR collateral and cannot be released without bank authorization");
        }

        StockReleaseRequest release = StockReleaseRequest.builder()
            .palletItemId(req.getPalletItemId())
            .tenantId(req.getTenantId())
            .warehouseId(pallet.getWarehouseId())
            .quantityToReleaseKg(req.getQuantityToReleaseKg())
            .releaseReason(req.getReleaseReason())
            .status(StockReleaseRequest.ReleaseStatus.otp_sent)
            .build();

        StockReleaseRequest saved = releaseRepo.save(release);

        // Send 6-digit OTP to farmer's phone
        otpService.sendOTP(req.getTenantId().toString(), "stock_release", saved.getId().toString());

        log.info("Stock release initiated: releaseId={}, palletId={}", saved.getId(), req.getPalletItemId());
        return saved;
    }

    @Transactional
    public StockReleaseRequest authorizeReleaseWithOTP(UUID releaseId, String otp, UUID operatorId) {
        StockReleaseRequest release = releaseRepo.findById(releaseId)
            .orElseThrow(() -> new IllegalArgumentException("Release request not found"));

        if (release.getStatus() != StockReleaseRequest.ReleaseStatus.otp_sent) {
            throw new IllegalStateException("Release is not in otp_sent state");
        }

        boolean valid = otpService.verifyOTP(release.getTenantId().toString(), otp, "stock_release");
        if (!valid) {
            release.setStatus(StockReleaseRequest.ReleaseStatus.rejected);
            releaseRepo.save(release);
            throw new SecurityException("Invalid or expired OTP");
        }

        release.setStatus(StockReleaseRequest.ReleaseStatus.authorized);
        release.setAuthorizedAt(OffsetDateTime.now());
        release.setAuthorizedByOperator(operatorId);

        kafkaEventService.publishInventoryEvent("STOCK_RELEASE_AUTHORIZED", release);
        log.info("Stock release authorized: releaseId={}", releaseId);

        return releaseRepo.save(release);
    }

    @Transactional
    public StockReleaseRequest completeRelease(UUID releaseId) {
        StockReleaseRequest release = releaseRepo.findById(releaseId)
            .orElseThrow(() -> new IllegalArgumentException("Release request not found"));

        if (release.getStatus() != StockReleaseRequest.ReleaseStatus.authorized) {
            throw new IllegalStateException("Release must be authorized before completion");
        }

        PalletItem pallet = palletRepo.findById(release.getPalletItemId())
            .orElseThrow(() -> new IllegalArgumentException("Pallet not found"));

        pallet.setActualOutwardDate(OffsetDateTime.now());
        palletRepo.save(pallet);

        release.setStatus(StockReleaseRequest.ReleaseStatus.completed);
        release.setCompletedAt(OffsetDateTime.now());

        kafkaEventService.publishInventoryEvent("PALLET_OUTWARD", pallet);
        log.info("Pallet outward completed: palletId={}", pallet.getId());

        return releaseRepo.save(release);
    }

    // ─── Queries ───────────────────────────────────────────────────────────

    public List<PalletItem> getInventoryByWarehouse(UUID warehouseId) {
        return palletRepo.findByWarehouseIdAndActualOutwardDateIsNull(warehouseId);
    }

    public List<PalletItem> getInventoryByTenant(UUID tenantId) {
        return palletRepo.findByTenantId(tenantId);
    }

    public long getActiveInventoryCount(UUID warehouseId) {
        return palletRepo.countActiveByWarehouse(warehouseId);
    }
}

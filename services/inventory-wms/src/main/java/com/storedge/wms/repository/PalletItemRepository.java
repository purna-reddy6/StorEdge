package com.storedge.wms.repository;

import com.storedge.wms.model.PalletItem;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface PalletItemRepository extends JpaRepository<PalletItem, UUID> {

    List<PalletItem> findByBookingId(UUID bookingId);

    List<PalletItem> findByWarehouseId(UUID warehouseId);

    List<PalletItem> findByTenantId(UUID tenantId);

    List<PalletItem> findByWarehouseIdAndActualOutwardDateIsNull(UUID warehouseId);

    Optional<PalletItem> findByRfidTagId(String rfidTagId);

    @Query("""
        SELECT p FROM PalletItem p
        WHERE p.warehouseId = :warehouseId
          AND p.slotPosition = :slot
          AND p.actualOutwardDate IS NULL
        """)
    Optional<PalletItem> findActiveBySlot(@Param("warehouseId") UUID warehouseId,
                                           @Param("slot") String slotPosition);

    @Query("""
        SELECT COUNT(p) FROM PalletItem p
        WHERE p.warehouseId = :warehouseId
          AND p.actualOutwardDate IS NULL
        """)
    long countActiveByWarehouse(@Param("warehouseId") UUID warehouseId);

    @Query("""
        SELECT p FROM PalletItem p
        WHERE p.enwrsPledged = true
          AND p.enwrsReceiptId = :receiptId
        """)
    Optional<PalletItem> findByEnwrsReceiptId(@Param("receiptId") String receiptId);
}

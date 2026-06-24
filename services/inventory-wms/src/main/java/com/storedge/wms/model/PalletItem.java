package com.storedge.wms.model;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.hibernate.annotations.CreationTimestamp;
import org.hibernate.annotations.UpdateTimestamp;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.time.OffsetDateTime;
import java.util.UUID;

@Entity
@Table(name = "pallet_items")
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class PalletItem {

    @Id
    @GeneratedValue(strategy = GenerationType.AUTO)
    private UUID id;

    @Column(name = "booking_id", nullable = false)
    private UUID bookingId;

    @Column(name = "warehouse_id", nullable = false)
    private UUID warehouseId;

    @Column(name = "tenant_id", nullable = false)
    private UUID tenantId;

    @Enumerated(EnumType.STRING)
    @Column(name = "commodity_type", nullable = false)
    private CommodityType commodityType;

    @Column(name = "commodity_description", length = 500)
    private String commodityDescription;

    @Column(name = "weight_kg", nullable = false, precision = 10, scale = 2)
    private BigDecimal weightKg;

    @Column(name = "volume_cubic_meters", precision = 8, scale = 3)
    private BigDecimal volumeCubicMeters;

    @Column(name = "bag_count")
    private Integer bagCount;

    @Column(name = "slot_position", length = 20)
    private String slotPosition;  // "A-03-02"

    @Column(name = "rfid_tag_id", length = 100)
    private String rfidTagId;

    @Column(name = "enwrs_pledged")
    private boolean enwrsPledged = false;

    @Column(name = "enwrs_receipt_id", length = 100)
    private String enwrsReceiptId;

    @Column(name = "inward_date")
    private OffsetDateTime inwardDate;

    @Column(name = "expected_outward_date")
    private LocalDate expectedOutwardDate;

    @Column(name = "actual_outward_date")
    private OffsetDateTime actualOutwardDate;

    @CreationTimestamp
    @Column(name = "created_at", updatable = false)
    private OffsetDateTime createdAt;

    @UpdateTimestamp
    @Column(name = "updated_at")
    private OffsetDateTime updatedAt;

    public enum CommodityType {
        potato, wheat, paddy, onion, garlic,
        fruits, vegetables, pulses, oilseeds, cereals,
        pharma, dairy,
        apparel, electronics, fmcg, auto_parts,
        other
    }
}

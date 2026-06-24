package com.storedge.wms.model;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.hibernate.annotations.CreationTimestamp;
import org.hibernate.annotations.UpdateTimestamp;

import java.math.BigDecimal;
import java.time.OffsetDateTime;
import java.util.UUID;

@Entity
@Table(name = "stock_release_requests")
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class StockReleaseRequest {

    @Id
    @GeneratedValue(strategy = GenerationType.AUTO)
    private UUID id;

    @Column(name = "pallet_item_id", nullable = false)
    private UUID palletItemId;

    @Column(name = "tenant_id", nullable = false)
    private UUID tenantId;

    @Column(name = "warehouse_id", nullable = false)
    private UUID warehouseId;

    @Column(name = "quantity_to_release_kg", nullable = false, precision = 10, scale = 2)
    private BigDecimal quantityToReleaseKg;

    @Column(name = "release_reason", length = 255)
    private String releaseReason;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    private ReleaseStatus status = ReleaseStatus.pending_otp;

    @Column(name = "otp_request_id")
    private UUID otpRequestId;

    @Column(name = "authorized_at")
    private OffsetDateTime authorizedAt;

    @Column(name = "authorized_by_operator")
    private UUID authorizedByOperator;

    @Column(name = "completed_at")
    private OffsetDateTime completedAt;

    @CreationTimestamp
    @Column(name = "created_at", updatable = false)
    private OffsetDateTime createdAt;

    @UpdateTimestamp
    @Column(name = "updated_at")
    private OffsetDateTime updatedAt;

    public enum ReleaseStatus {
        pending_otp, otp_sent, authorized, rejected, completed
    }
}

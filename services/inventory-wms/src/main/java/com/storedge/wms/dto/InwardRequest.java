package com.storedge.wms.dto;

import jakarta.validation.constraints.*;
import lombok.Data;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.util.UUID;

@Data
public class InwardRequest {

    @NotNull
    private UUID bookingId;

    @NotNull
    private UUID warehouseId;

    @NotNull
    private UUID tenantId;

    @NotBlank
    private String commodityType;

    private String commodityDescription;

    @NotNull
    @DecimalMin("0.01")
    private BigDecimal weightKg;

    private BigDecimal volumeCubicMeters;

    private Integer bagCount;

    private String slotPosition;  // Assigned by WMS slotting engine if null

    private String rfidTagId;

    private LocalDate expectedOutwardDate;
}

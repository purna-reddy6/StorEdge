package com.storedge.wms.dto;

import jakarta.validation.constraints.*;
import lombok.Data;

import java.math.BigDecimal;
import java.util.UUID;

@Data
public class ReleaseRequestDto {

    @NotNull
    private UUID palletItemId;

    @NotNull
    private UUID tenantId;

    @NotNull
    @DecimalMin("0.01")
    private BigDecimal quantityToReleaseKg;

    private String releaseReason;
}

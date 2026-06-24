package com.storedge.wms.model;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.hibernate.annotations.UpdateTimestamp;

import java.time.OffsetDateTime;
import java.util.UUID;

@Entity
@Table(name = "warehouse_floor_plans")
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class WarehouseFloorPlan {

    @Id
    @GeneratedValue(strategy = GenerationType.AUTO)
    private UUID id;

    @Column(name = "warehouse_id", nullable = false, unique = true)
    private UUID warehouseId;

    @Column(name = "grid_rows", nullable = false)
    private Integer gridRows;

    @Column(name = "grid_columns", nullable = false)
    private Integer gridColumns;

    @Column(name = "grid_data", columnDefinition = "jsonb", nullable = false)
    private String gridData = "{}";  // JSON slot occupancy map

    @Column(name = "svg_layout_url", length = 500)
    private String svgLayoutUrl;

    @UpdateTimestamp
    @Column(name = "updated_at")
    private OffsetDateTime updatedAt;
}

package com.storedge.wms.repository;

import com.storedge.wms.model.WarehouseFloorPlan;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.Optional;
import java.util.UUID;

@Repository
public interface FloorPlanRepository extends JpaRepository<WarehouseFloorPlan, UUID> {
    Optional<WarehouseFloorPlan> findByWarehouseId(UUID warehouseId);
}

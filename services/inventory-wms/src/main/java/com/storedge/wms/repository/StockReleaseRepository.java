package com.storedge.wms.repository;

import com.storedge.wms.model.StockReleaseRequest;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.UUID;

@Repository
public interface StockReleaseRepository extends JpaRepository<StockReleaseRequest, UUID> {

    List<StockReleaseRequest> findByTenantId(UUID tenantId);

    List<StockReleaseRequest> findByPalletItemId(UUID palletItemId);

    List<StockReleaseRequest> findByStatus(StockReleaseRequest.ReleaseStatus status);
}

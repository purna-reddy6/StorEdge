package com.storedge.wms.service;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Service;

import java.time.OffsetDateTime;
import java.util.HashMap;
import java.util.Map;

@Service
@RequiredArgsConstructor
@Slf4j
public class KafkaEventService {

    private final KafkaTemplate<String, Object> kafkaTemplate;

    @Value("${storedge.kafka.topics.inventory-events:storedge.inventory.events}")
    private String inventoryTopic;

    public void publishInventoryEvent(String eventType, Object payload) {
        Map<String, Object> event = new HashMap<>();
        event.put("event_type", eventType);
        event.put("timestamp", OffsetDateTime.now().toString());
        event.put("payload", payload);

        kafkaTemplate.send(inventoryTopic, eventType, event)
            .whenComplete((result, ex) -> {
                if (ex != null) {
                    log.error("Failed to publish Kafka event: type={}", eventType, ex);
                } else {
                    log.debug("Published Kafka event: type={}", eventType);
                }
            });
    }
}

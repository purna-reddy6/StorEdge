package com.storedge.wms;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.scheduling.annotation.EnableScheduling;

@SpringBootApplication
@EnableScheduling
public class InventoryWmsApplication {
    public static void main(String[] args) {
        SpringApplication.run(InventoryWmsApplication.class, args);
    }
}

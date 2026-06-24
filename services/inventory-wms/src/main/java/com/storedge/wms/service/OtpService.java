package com.storedge.wms.service;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.stereotype.Service;

import java.security.SecureRandom;
import java.time.OffsetDateTime;

/**
 * OtpService generates and verifies 6-digit OTPs for stock release authorization.
 * Implements the "travel tax" elimination described in the blueprint (Part 6, Part 15):
 * farmers can authorize crop releases remotely without 45km round trips.
 */
@Service
@RequiredArgsConstructor
@Slf4j
public class OtpService {

    private final JdbcTemplate jdbc;
    private final SecureRandom secureRandom = new SecureRandom();

    public String sendOTP(String userId, String purpose, String referenceId) {
        String otp = generateOTP();
        String otpHash = hashOTP(otp);
        OffsetDateTime expiresAt = OffsetDateTime.now().plusMinutes(10);

        jdbc.update("""
            INSERT INTO otp_requests (phone, otp_hash, purpose, expires_at)
            SELECT phone, ?, ?, ?
            FROM users WHERE id = ?::uuid
            """,
            otpHash, purpose + ":" + referenceId, expiresAt, userId
        );

        // Production: dispatch to SMS gateway (Twilio/Gupshup)
        log.info("OTP generated for userId={} purpose={} [dev: {}]", userId, purpose, otp);

        return otp; // returned for dev/test only
    }

    public boolean verifyOTP(String userId, String otp, String purpose) {
        String otpHash = hashOTP(otp);

        Integer count = jdbc.queryForObject("""
            SELECT COUNT(*) FROM otp_requests
            WHERE phone = (SELECT phone FROM users WHERE id = ?::uuid)
              AND otp_hash = ?
              AND purpose LIKE ?
              AND expires_at > NOW()
              AND used_at IS NULL
            """,
            Integer.class, userId, otpHash, purpose + "%"
        );

        if (count != null && count > 0) {
            jdbc.update("""
                UPDATE otp_requests SET used_at = NOW()
                WHERE phone = (SELECT phone FROM users WHERE id = ?::uuid)
                  AND otp_hash = ?
                  AND used_at IS NULL
                """,
                userId, otpHash
            );
            return true;
        }
        return false;
    }

    private String generateOTP() {
        return String.format("%06d", secureRandom.nextInt(1_000_000));
    }

    private String hashOTP(String otp) {
        // Simple hex encoding for MVP — SHA-256 in production
        return Integer.toHexString(otp.hashCode() ^ "storedge".hashCode());
    }
}

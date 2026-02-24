package com.flow.ledger.dto;

import jakarta.validation.constraints.NotNull;

import java.math.BigDecimal;

public record AdjustBalanceRequest(
        @NotNull(message = "New balance is required")
        BigDecimal newBalance
) {
}

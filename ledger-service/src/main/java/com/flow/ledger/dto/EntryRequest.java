package com.flow.ledger.dto;

import com.flow.ledger.model.entity.EntryType;
import jakarta.validation.constraints.DecimalMin;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Positive;

import java.math.BigDecimal;

public record EntryRequest(
        @NotNull(message = "Account ID is required")
        @Positive
        Long accountId,

        @NotNull(message = "Amount is required")
        @DecimalMin(value = "0.0001", message = "Amount must be positive")
        BigDecimal amount,

        @NotNull(message = "Entry type (DEBIT/CREDIT) is required")
        EntryType type
) {
}

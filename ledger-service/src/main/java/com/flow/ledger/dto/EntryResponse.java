package com.flow.ledger.dto;

import com.flow.ledger.model.entity.EntryType;

import java.math.BigDecimal;

public record EntryResponse(
        Long id,
        Long accountId,
        BigDecimal amount,
        EntryType type
) {
}

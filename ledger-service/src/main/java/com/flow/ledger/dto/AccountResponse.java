package com.flow.ledger.dto;

import com.flow.ledger.model.entity.AccountType;

import java.math.BigDecimal;

public record AccountResponse(
        Long id,
        String code,
        String name,
        AccountType type,
        BigDecimal balance,
        String color
) {
}

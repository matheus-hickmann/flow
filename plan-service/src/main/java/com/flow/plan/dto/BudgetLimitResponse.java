package com.flow.plan.dto;

import com.flow.plan.model.entity.LimitType;

import java.math.BigDecimal;
import java.time.Instant;

public record BudgetLimitResponse(
        Long id,
        String category,
        LimitType limitType,
        BigDecimal limitValue,
        Instant createdAt
) {
}

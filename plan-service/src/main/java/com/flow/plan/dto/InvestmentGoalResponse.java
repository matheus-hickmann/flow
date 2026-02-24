package com.flow.plan.dto;

import java.math.BigDecimal;
import java.time.Instant;

public record InvestmentGoalResponse(
        Long id,
        String name,
        BigDecimal expectedReturnRate,
        BigDecimal monthlyContribution,
        BigDecimal targetAmount,
        Instant createdAt
) {
}

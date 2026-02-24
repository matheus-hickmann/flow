package com.flow.plan.dto;

import jakarta.validation.constraints.DecimalMin;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Size;

import java.math.BigDecimal;

public record InvestmentGoalRequest(
        @NotBlank(message = "Name is required")
        @Size(max = 255)
        String name,

        @NotNull(message = "Expected return rate is required")
        @DecimalMin(value = "0", message = "Expected return rate must be non-negative")
        BigDecimal expectedReturnRate,

        @NotNull(message = "Monthly contribution is required")
        @DecimalMin(value = "0", message = "Monthly contribution must be non-negative")
        BigDecimal monthlyContribution,

        @DecimalMin(value = "0", message = "Target amount must be non-negative")
        BigDecimal targetAmount
) {
}

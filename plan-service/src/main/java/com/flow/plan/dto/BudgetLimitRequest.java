package com.flow.plan.dto;

import com.flow.plan.model.entity.LimitType;
import jakarta.validation.constraints.DecimalMin;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Size;

import java.math.BigDecimal;

public record BudgetLimitRequest(
        @NotBlank(message = "Category is required")
        @Size(max = 100)
        String category,

        @NotNull(message = "Limit type (ABSOLUTE or PERCENTAGE) is required")
        LimitType limitType,

        @NotNull(message = "Limit value is required")
        @DecimalMin(value = "0", inclusive = false, message = "Limit value must be positive")
        BigDecimal limitValue
) {
}

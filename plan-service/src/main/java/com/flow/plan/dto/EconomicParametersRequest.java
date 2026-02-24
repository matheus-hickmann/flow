package com.flow.plan.dto;

import jakarta.validation.constraints.DecimalMin;
import jakarta.validation.constraints.NotNull;

import java.math.BigDecimal;

public record EconomicParametersRequest(
        @NotNull(message = "Selic rate is required")
        @DecimalMin(value = "0", message = "Selic rate must be non-negative")
        BigDecimal selicRate,

        @NotNull(message = "IPCA rate is required")
        @DecimalMin(value = "0", message = "IPCA rate must be non-negative")
        BigDecimal ipcaRate
) {
}

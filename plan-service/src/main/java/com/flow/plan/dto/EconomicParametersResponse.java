package com.flow.plan.dto;

import java.math.BigDecimal;
import java.time.Instant;

public record EconomicParametersResponse(
        Long id,
        BigDecimal selicRate,
        BigDecimal ipcaRate,
        Instant updatedAt
) {
}

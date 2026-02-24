package com.flow.ledger.dto;

import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Pattern;
import jakarta.validation.constraints.Size;

import java.math.BigDecimal;

public record CreateAccountRequest(
        @NotNull(message = "Name is required")
        @Size(min = 1, max = 255)
        String name,

        @NotNull(message = "Initial balance is required")
        BigDecimal initialBalance,

        @Size(max = 7)
        @Pattern(regexp = "^#[0-9A-Fa-f]{6}$", message = "Color must be a hex value e.g. #3b82f6")
        String color
) {
    public String colorOrDefault() {
        return color != null && !color.isBlank() ? color : "#3b82f6";
    }
}

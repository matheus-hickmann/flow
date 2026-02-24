package com.flow.ledger.dto;

import jakarta.validation.Valid;
import jakarta.validation.constraints.NotEmpty;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Size;

import java.util.List;

public record PostTransactionRequest(
        @NotNull(message = "Description is required")
        @Size(min = 1, max = 500)
        String description,

        @Size(max = 100)
        String referenceId,

        @NotNull(message = "Entries are required")
        @NotEmpty(message = "At least two entries are required for double-entry")
        @Valid
        List<EntryRequest> entries
) {
}

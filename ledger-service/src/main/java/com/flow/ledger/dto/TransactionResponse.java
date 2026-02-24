package com.flow.ledger.dto;

import java.time.Instant;
import java.util.List;

public record TransactionResponse(
        Long id,
        String description,
        Instant timestamp,
        String referenceId,
        List<EntryResponse> entries
) {
}

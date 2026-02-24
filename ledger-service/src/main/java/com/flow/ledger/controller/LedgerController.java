package com.flow.ledger.controller;

import com.flow.ledger.dto.PostTransactionRequest;
import com.flow.ledger.dto.TransactionResponse;
import com.flow.ledger.service.LedgerService;
import jakarta.validation.Valid;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping(path = "/api/v1/ledger/transactions", produces = MediaType.APPLICATION_JSON_VALUE)
public class LedgerController {

    private final LedgerService ledgerService;

    public LedgerController(LedgerService ledgerService) {
        this.ledgerService = ledgerService;
    }

    @PostMapping(consumes = MediaType.APPLICATION_JSON_VALUE)
    @ResponseStatus(HttpStatus.CREATED)
    public TransactionResponse postTransaction(@Valid @RequestBody PostTransactionRequest request) {
        return ledgerService.postTransaction(request);
    }
}

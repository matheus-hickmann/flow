package com.flow.ledger.controller;

import com.flow.ledger.dto.AccountResponse;
import com.flow.ledger.dto.AdjustBalanceRequest;
import com.flow.ledger.dto.CreateAccountRequest;
import com.flow.ledger.service.AccountService;
import jakarta.validation.Valid;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PatchMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
@RequestMapping(path = "/api/v1/ledger/accounts", produces = MediaType.APPLICATION_JSON_VALUE)
public class AccountController {

    private final AccountService accountService;

    public AccountController(AccountService accountService) {
        this.accountService = accountService;
    }

    @GetMapping
    public List<AccountResponse> list() {
        return accountService.listAll();
    }

    @PostMapping(consumes = MediaType.APPLICATION_JSON_VALUE)
    @ResponseStatus(HttpStatus.CREATED)
    public AccountResponse create(@Valid @RequestBody CreateAccountRequest request) {
        return accountService.create(request);
    }

    @PatchMapping(value = "/{id}/balance", consumes = MediaType.APPLICATION_JSON_VALUE)
    public AccountResponse adjustBalance(@PathVariable Long id, @Valid @RequestBody AdjustBalanceRequest request) {
        return accountService.adjustBalance(id, request);
    }
}

package com.flow.plan.controller;

import com.flow.plan.dto.BudgetLimitRequest;
import com.flow.plan.dto.BudgetLimitResponse;
import com.flow.plan.service.BudgetLimitService;
import jakarta.validation.Valid;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
@RequestMapping(path = "/api/v1/budgets", produces = MediaType.APPLICATION_JSON_VALUE)
public class BudgetLimitController {

    private final BudgetLimitService service;

    public BudgetLimitController(BudgetLimitService service) {
        this.service = service;
    }

    @PostMapping(consumes = MediaType.APPLICATION_JSON_VALUE)
    @ResponseStatus(HttpStatus.CREATED)
    public BudgetLimitResponse create(@Valid @RequestBody BudgetLimitRequest request) {
        return service.create(request);
    }

    @GetMapping
    public List<BudgetLimitResponse> findAll() {
        return service.findAll();
    }

    @GetMapping("/{id}")
    public BudgetLimitResponse findById(@PathVariable Long id) {
        return service.findById(id);
    }

    @PutMapping(value = "/{id}", consumes = MediaType.APPLICATION_JSON_VALUE)
    public BudgetLimitResponse update(@PathVariable Long id, @Valid @RequestBody BudgetLimitRequest request) {
        return service.update(id, request);
    }
}

package com.flow.plan.controller;

import com.flow.plan.dto.EconomicParametersRequest;
import com.flow.plan.dto.EconomicParametersResponse;
import com.flow.plan.service.EconomicParametersService;
import jakarta.validation.Valid;
import org.springframework.http.MediaType;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping(path = "/api/v1/economic-parameters", produces = MediaType.APPLICATION_JSON_VALUE)
public class EconomicParametersController {

    private final EconomicParametersService service;

    public EconomicParametersController(EconomicParametersService service) {
        this.service = service;
    }

    @GetMapping
    public EconomicParametersResponse get() {
        return service.get();
    }

    @PutMapping(consumes = MediaType.APPLICATION_JSON_VALUE)
    public EconomicParametersResponse createOrUpdate(@Valid @RequestBody EconomicParametersRequest request) {
        return service.createOrUpdate(request);
    }
}

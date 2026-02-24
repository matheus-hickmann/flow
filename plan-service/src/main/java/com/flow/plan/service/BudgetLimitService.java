package com.flow.plan.service;

import com.flow.plan.dto.BudgetLimitRequest;
import com.flow.plan.dto.BudgetLimitResponse;
import com.flow.plan.exception.ResourceNotFoundException;
import com.flow.plan.model.entity.BudgetLimit;
import com.flow.plan.model.entity.LimitType;
import com.flow.plan.repository.BudgetLimitRepository;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;

@Service
public class BudgetLimitService {

    private final BudgetLimitRepository repository;

    public BudgetLimitService(BudgetLimitRepository repository) {
        this.repository = repository;
    }

    @Transactional
    public BudgetLimitResponse create(BudgetLimitRequest request) {
        BudgetLimit entity = new BudgetLimit(
                request.category(),
                request.limitType(),
                request.limitValue()
        );
        BudgetLimit saved = repository.save(entity);
        return toResponse(saved);
    }

    public List<BudgetLimitResponse> findAll() {
        return repository.findAllByOrderByCategoryAsc().stream()
                .map(BudgetLimitService::toResponse)
                .toList();
    }

    public BudgetLimitResponse findById(Long id) {
        BudgetLimit entity = repository.findById(id)
                .orElseThrow(() -> new ResourceNotFoundException("BudgetLimit", id));
        return toResponse(entity);
    }

    @Transactional
    public BudgetLimitResponse update(Long id, BudgetLimitRequest request) {
        BudgetLimit entity = repository.findById(id)
                .orElseThrow(() -> new ResourceNotFoundException("BudgetLimit", id));
        entity.setCategory(request.category());
        entity.setLimitType(request.limitType());
        entity.setLimitValue(request.limitValue());
        BudgetLimit saved = repository.save(entity);
        return toResponse(saved);
    }

    private static BudgetLimitResponse toResponse(BudgetLimit entity) {
        return new BudgetLimitResponse(
                entity.getId(),
                entity.getCategory(),
                entity.getLimitType(),
                entity.getLimitValue(),
                entity.getCreatedAt()
        );
    }
}

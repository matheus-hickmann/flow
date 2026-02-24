package com.flow.plan.service;

import com.flow.plan.dto.InvestmentGoalRequest;
import com.flow.plan.dto.InvestmentGoalResponse;
import com.flow.plan.exception.ResourceNotFoundException;
import com.flow.plan.model.entity.InvestmentGoal;
import com.flow.plan.repository.InvestmentGoalRepository;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;

@Service
public class InvestmentGoalService {

    private final InvestmentGoalRepository repository;

    public InvestmentGoalService(InvestmentGoalRepository repository) {
        this.repository = repository;
    }

    @Transactional
    public InvestmentGoalResponse create(InvestmentGoalRequest request) {
        InvestmentGoal entity = new InvestmentGoal(
                request.name(),
                request.expectedReturnRate(),
                request.monthlyContribution(),
                request.targetAmount()
        );
        InvestmentGoal saved = repository.save(entity);
        return toResponse(saved);
    }

    public List<InvestmentGoalResponse> findAll() {
        return repository.findAllByOrderByCreatedAtDesc().stream()
                .map(InvestmentGoalService::toResponse)
                .toList();
    }

    public InvestmentGoalResponse findById(Long id) {
        InvestmentGoal entity = repository.findById(id)
                .orElseThrow(() -> new ResourceNotFoundException("InvestmentGoal", id));
        return toResponse(entity);
    }

    @Transactional
    public InvestmentGoalResponse update(Long id, InvestmentGoalRequest request) {
        InvestmentGoal entity = repository.findById(id)
                .orElseThrow(() -> new ResourceNotFoundException("InvestmentGoal", id));
        entity.setName(request.name());
        entity.setExpectedReturnRate(request.expectedReturnRate());
        entity.setMonthlyContribution(request.monthlyContribution());
        entity.setTargetAmount(request.targetAmount());
        InvestmentGoal saved = repository.save(entity);
        return toResponse(saved);
    }

    private static InvestmentGoalResponse toResponse(InvestmentGoal entity) {
        return new InvestmentGoalResponse(
                entity.getId(),
                entity.getName(),
                entity.getExpectedReturnRate(),
                entity.getMonthlyContribution(),
                entity.getTargetAmount(),
                entity.getCreatedAt()
        );
    }
}

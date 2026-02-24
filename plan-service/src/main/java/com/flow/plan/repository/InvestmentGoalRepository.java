package com.flow.plan.repository;

import com.flow.plan.model.entity.InvestmentGoal;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.List;

public interface InvestmentGoalRepository extends JpaRepository<InvestmentGoal, Long> {

    List<InvestmentGoal> findAllByOrderByCreatedAtDesc();
}

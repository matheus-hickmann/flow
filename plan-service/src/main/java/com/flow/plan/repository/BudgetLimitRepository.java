package com.flow.plan.repository;

import com.flow.plan.model.entity.BudgetLimit;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.List;

public interface BudgetLimitRepository extends JpaRepository<BudgetLimit, Long> {

    List<BudgetLimit> findAllByOrderByCategoryAsc();
}

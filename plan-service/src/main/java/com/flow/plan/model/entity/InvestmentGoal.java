package com.flow.plan.model.entity;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.Table;

import java.math.BigDecimal;
import java.time.Instant;
import java.util.Objects;

@Entity
@Table(name = "investment_goal")
public class InvestmentGoal {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(nullable = false, length = 255)
    private String name;

    @Column(name = "expected_return_rate", nullable = false, precision = 10, scale = 4)
    private BigDecimal expectedReturnRate;

    @Column(name = "monthly_contribution", nullable = false, precision = 19, scale = 4)
    private BigDecimal monthlyContribution;

    @Column(name = "target_amount", precision = 19, scale = 4)
    private BigDecimal targetAmount;

    @Column(name = "created_at", nullable = false)
    private Instant createdAt;

    protected InvestmentGoal() {
    }

    public InvestmentGoal(String name, BigDecimal expectedReturnRate, BigDecimal monthlyContribution, BigDecimal targetAmount) {
        this.name = name;
        this.expectedReturnRate = expectedReturnRate;
        this.monthlyContribution = monthlyContribution;
        this.targetAmount = targetAmount;
        this.createdAt = Instant.now();
    }

    public Long getId() {
        return id;
    }

    public String getName() {
        return name;
    }

    public BigDecimal getExpectedReturnRate() {
        return expectedReturnRate;
    }

    public BigDecimal getMonthlyContribution() {
        return monthlyContribution;
    }

    public BigDecimal getTargetAmount() {
        return targetAmount;
    }

    public Instant getCreatedAt() {
        return createdAt;
    }

    public void setName(String name) {
        this.name = name;
    }

    public void setExpectedReturnRate(BigDecimal expectedReturnRate) {
        this.expectedReturnRate = expectedReturnRate;
    }

    public void setMonthlyContribution(BigDecimal monthlyContribution) {
        this.monthlyContribution = monthlyContribution;
    }

    public void setTargetAmount(BigDecimal targetAmount) {
        this.targetAmount = targetAmount;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        InvestmentGoal that = (InvestmentGoal) o;
        return id != null && Objects.equals(id, that.id);
    }

    @Override
    public int hashCode() {
        return getClass().hashCode();
    }
}

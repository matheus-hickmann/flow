package com.flow.plan.model.entity;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.EnumType;
import jakarta.persistence.Enumerated;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.Table;

import java.math.BigDecimal;
import java.time.Instant;
import java.util.Objects;

@Entity
@Table(name = "budget_limit")
public class BudgetLimit {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(nullable = false, length = 100)
    private String category;

    @Enumerated(EnumType.STRING)
    @Column(name = "limit_type", nullable = false, length = 20)
    private LimitType limitType;

    @Column(name = "limit_value", nullable = false, precision = 19, scale = 4)
    private BigDecimal limitValue;

    @Column(name = "created_at", nullable = false)
    private Instant createdAt;

    protected BudgetLimit() {
    }

    public BudgetLimit(String category, LimitType limitType, BigDecimal limitValue) {
        this.category = category;
        this.limitType = limitType;
        this.limitValue = limitValue;
        this.createdAt = Instant.now();
    }

    public Long getId() {
        return id;
    }

    public String getCategory() {
        return category;
    }

    public LimitType getLimitType() {
        return limitType;
    }

    public BigDecimal getLimitValue() {
        return limitValue;
    }

    public Instant getCreatedAt() {
        return createdAt;
    }

    public void setCategory(String category) {
        this.category = category;
    }

    public void setLimitType(LimitType limitType) {
        this.limitType = limitType;
    }

    public void setLimitValue(BigDecimal limitValue) {
        this.limitValue = limitValue;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        BudgetLimit that = (BudgetLimit) o;
        return id != null && Objects.equals(id, that.id);
    }

    @Override
    public int hashCode() {
        return getClass().hashCode();
    }
}

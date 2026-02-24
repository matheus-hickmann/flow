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
@Table(name = "economic_parameters")
public class EconomicParameters {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(name = "selic_rate", nullable = false, precision = 10, scale = 4)
    private BigDecimal selicRate;

    @Column(name = "ipca_rate", nullable = false, precision = 10, scale = 4)
    private BigDecimal ipcaRate;

    @Column(name = "updated_at", nullable = false)
    private Instant updatedAt;

    protected EconomicParameters() {
    }

    public EconomicParameters(BigDecimal selicRate, BigDecimal ipcaRate) {
        this.selicRate = selicRate;
        this.ipcaRate = ipcaRate;
        this.updatedAt = Instant.now();
    }

    public Long getId() {
        return id;
    }

    public BigDecimal getSelicRate() {
        return selicRate;
    }

    public BigDecimal getIpcaRate() {
        return ipcaRate;
    }

    public Instant getUpdatedAt() {
        return updatedAt;
    }

    public void setSelicRate(BigDecimal selicRate) {
        this.selicRate = selicRate;
        this.updatedAt = Instant.now();
    }

    public void setIpcaRate(BigDecimal ipcaRate) {
        this.ipcaRate = ipcaRate;
        this.updatedAt = Instant.now();
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        EconomicParameters that = (EconomicParameters) o;
        return id != null && Objects.equals(id, that.id);
    }

    @Override
    public int hashCode() {
        return getClass().hashCode();
    }
}

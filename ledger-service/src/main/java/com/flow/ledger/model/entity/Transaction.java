package com.flow.ledger.model.entity;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.OneToMany;
import jakarta.persistence.Table;

import java.time.Instant;
import java.util.ArrayList;
import java.util.List;
import java.util.Objects;

@Entity
@Table(name = "ledger_transaction")
public class Transaction {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(nullable = false, length = 500)
    private String description;

    @Column(nullable = false)
    private Instant timestamp;

    @Column(name = "reference_id", length = 100)
    private String referenceId;

    @OneToMany(mappedBy = "transaction", orphanRemoval = true, cascade = jakarta.persistence.CascadeType.PERSIST)
    private List<Entry> entries = new ArrayList<>();

    protected Transaction() {
    }

    public Transaction(String description, String referenceId) {
        this.description = description;
        this.referenceId = referenceId;
        this.timestamp = Instant.now();
    }

    public Long getId() {
        return id;
    }

    public String getDescription() {
        return description;
    }

    public Instant getTimestamp() {
        return timestamp;
    }

    public String getReferenceId() {
        return referenceId;
    }

    public List<Entry> getEntries() {
        return entries;
    }

    public void addEntry(Entry entry) {
        this.entries.add(entry);
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        Transaction that = (Transaction) o;
        return id != null && Objects.equals(id, that.id);
    }

    @Override
    public int hashCode() {
        return getClass().hashCode();
    }
}

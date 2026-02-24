package com.flow.ledger.service;

import com.flow.ledger.dto.EntryRequest;
import com.flow.ledger.dto.PostTransactionRequest;
import com.flow.ledger.dto.TransactionResponse;
import com.flow.ledger.exception.AccountNotFoundException;
import com.flow.ledger.exception.InvalidLedgerTransactionException;
import com.flow.ledger.model.entity.Account;
import com.flow.ledger.model.entity.AccountType;
import com.flow.ledger.model.entity.Entry;
import com.flow.ledger.model.entity.EntryType;
import com.flow.ledger.model.entity.Transaction;
import com.flow.ledger.repository.AccountRepository;
import com.flow.ledger.repository.TransactionRepository;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.stream.Collectors;

@Service
public class LedgerService {

    private static final String DEBITS_MUST_EQUAL_CREDITS = "Sum of debits must equal sum of credits for double-entry bookkeeping";

    private final AccountRepository accountRepository;
    private final TransactionRepository transactionRepository;

    public LedgerService(AccountRepository accountRepository, TransactionRepository transactionRepository) {
        this.accountRepository = accountRepository;
        this.transactionRepository = transactionRepository;
    }

    @Transactional
    public TransactionResponse postTransaction(PostTransactionRequest request) {
        List<EntryRequest> entries = request.entries();
        if (entries.size() < 2) {
            throw new InvalidLedgerTransactionException("At least two entries are required for double-entry bookkeeping");
        }

        BigDecimal totalDebits = entries.stream()
                .filter(e -> e.type() == EntryType.DEBIT)
                .map(EntryRequest::amount)
                .reduce(BigDecimal.ZERO, BigDecimal::add);
        BigDecimal totalCredits = entries.stream()
                .filter(e -> e.type() == EntryType.CREDIT)
                .map(EntryRequest::amount)
                .reduce(BigDecimal.ZERO, BigDecimal::add);

        if (totalDebits.compareTo(totalCredits) != 0) {
            throw new InvalidLedgerTransactionException(DEBITS_MUST_EQUAL_CREDITS);
        }

        Set<Long> accountIds = entries.stream().map(EntryRequest::accountId).collect(Collectors.toSet());
        Map<Long, Account> accountsById = loadAccountsWithOptimisticLock(accountIds);

        Transaction transaction = new Transaction(request.description(), request.referenceId());

        for (EntryRequest er : entries) {
            Account account = accountsById.get(er.accountId());
            if (account == null) {
                throw new AccountNotFoundException(er.accountId());
            }
            Entry entry = new Entry(transaction, account, er.amount(), er.type());
            transaction.addEntry(entry);
            updateAccountBalance(account, er.amount(), er.type());
        }

        transactionRepository.save(transaction);

        return toResponse(transaction);
    }

    private Map<Long, Account> loadAccountsWithOptimisticLock(Set<Long> accountIds) {
        List<Account> accounts = accountIds.stream()
                .map(accountRepository::findByIdWithOptimisticLock)
                .filter(opt -> opt.isPresent())
                .map(opt -> opt.orElseThrow())
                .toList();
        if (accounts.size() != accountIds.size()) {
            Set<Long> foundIds = accounts.stream().map(Account::getId).collect(Collectors.toSet());
            Long missing = accountIds.stream().filter(id -> !foundIds.contains(id)).findFirst().orElse(null);
            throw new AccountNotFoundException(missing);
        }
        return accounts.stream().collect(Collectors.toMap(Account::getId, a -> a));
    }

    private void updateAccountBalance(Account account, BigDecimal amount, EntryType entryType) {
        AccountType accountType = account.getType();
        BigDecimal current = account.getBalance();
        BigDecimal delta;
        if (entryType == EntryType.DEBIT) {
            delta = (accountType == AccountType.ASSET || accountType == AccountType.EXPENSE) ? amount : amount.negate();
        } else {
            delta = (accountType == AccountType.LIABILITY || accountType == AccountType.REVENUE) ? amount : amount.negate();
        }
        account.setBalance(current.add(delta));
    }

    private static TransactionResponse toResponse(Transaction transaction) {
        List<com.flow.ledger.dto.EntryResponse> entryResponses = transaction.getEntries().stream()
                .map(e -> new com.flow.ledger.dto.EntryResponse(
                        e.getId(),
                        e.getAccount().getId(),
                        e.getAmount(),
                        e.getType()))
                .toList();
        return new TransactionResponse(
                transaction.getId(),
                transaction.getDescription(),
                transaction.getTimestamp(),
                transaction.getReferenceId(),
                entryResponses);
    }
}

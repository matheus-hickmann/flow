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
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.mockito.stubbing.Answer;

import java.math.BigDecimal;
import java.util.List;
import java.util.Optional;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.Mockito.any;
import static org.mockito.Mockito.doAnswer;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class LedgerServiceTest {

    @Mock
    private AccountRepository accountRepository;

    @Mock
    private TransactionRepository transactionRepository;

    @InjectMocks
    private LedgerService ledgerService;

    private Account assetAccount;
    private Account liabilityAccount;

    @BeforeEach
    void setUp() {
        assetAccount = new Account("CASH", "Cash", AccountType.ASSET);
        setAccountId(assetAccount, 1L);
        liabilityAccount = new Account("PAYABLE", "Accounts Payable", AccountType.LIABILITY);
        setAccountId(liabilityAccount, 2L);
    }

    @Nested
    @DisplayName("postTransaction")
    class PostTransaction {

        @Test
        void postTransaction_whenDebitsEqualCredits_createsTransactionAndReturnsResponse() {
            BigDecimal amount = new BigDecimal("100.00");
            PostTransactionRequest request = new PostTransactionRequest(
                    "Payment received",
                    "REF-001",
                    List.of(
                            new EntryRequest(1L, amount, EntryType.DEBIT),
                            new EntryRequest(2L, amount, EntryType.CREDIT)
                    ));
            when(accountRepository.findByIdWithOptimisticLock(1L)).thenReturn(Optional.of(assetAccount));
            when(accountRepository.findByIdWithOptimisticLock(2L)).thenReturn(Optional.of(liabilityAccount));
            stubTransactionSaveWithIds();

            TransactionResponse response = ledgerService.postTransaction(request);

            assertThat(response).isNotNull();
            assertThat(response.description()).isEqualTo("Payment received");
            assertThat(response.referenceId()).isEqualTo("REF-001");
            assertThat(response.entries()).hasSize(2);
            assertThat(response.timestamp()).isNotNull();

            ArgumentCaptor<Transaction> txCaptor = ArgumentCaptor.forClass(Transaction.class);
            verify(transactionRepository).save(txCaptor.capture());
            Transaction saved = txCaptor.getValue();
            assertThat(saved.getEntries()).hasSize(2);
            assertThat(assetAccount.getBalance()).isEqualByComparingTo(amount);
            assertThat(liabilityAccount.getBalance()).isEqualByComparingTo(amount);
        }

        @Test
        void postTransaction_whenDebitsNotEqualCredits_throwsInvalidLedgerTransactionException() {
            PostTransactionRequest request = new PostTransactionRequest(
                    "Invalid",
                    null,
                    List.of(
                            new EntryRequest(1L, new BigDecimal("100"), EntryType.DEBIT),
                            new EntryRequest(2L, new BigDecimal("50"), EntryType.CREDIT)
                    ));

            assertThatThrownBy(() -> ledgerService.postTransaction(request))
                    .isInstanceOf(InvalidLedgerTransactionException.class)
                    .hasMessageContaining("Sum of debits must equal sum of credits");
        }

        @Test
        void postTransaction_whenLessThanTwoEntries_throwsInvalidLedgerTransactionException() {
            PostTransactionRequest request = new PostTransactionRequest(
                    "Single entry",
                    null,
                    List.of(new EntryRequest(1L, new BigDecimal("100"), EntryType.DEBIT)));

            assertThatThrownBy(() -> ledgerService.postTransaction(request))
                    .isInstanceOf(InvalidLedgerTransactionException.class)
                    .hasMessageContaining("At least two entries are required");
        }

        @Test
        void postTransaction_whenAccountNotFound_throwsAccountNotFoundException() {
            PostTransactionRequest request = new PostTransactionRequest(
                    "Transfer",
                    null,
                    List.of(
                            new EntryRequest(1L, new BigDecimal("100"), EntryType.DEBIT),
                            new EntryRequest(999L, new BigDecimal("100"), EntryType.CREDIT)
                    ));
            when(accountRepository.findByIdWithOptimisticLock(1L)).thenReturn(Optional.of(assetAccount));
            when(accountRepository.findByIdWithOptimisticLock(999L)).thenReturn(Optional.empty());

            assertThatThrownBy(() -> ledgerService.postTransaction(request))
                    .isInstanceOf(AccountNotFoundException.class)
                    .hasMessageContaining("999");
        }

        @Test
        void postTransaction_assetDebitIncreasesBalance_liabilityCreditIncreasesBalance() {
            BigDecimal amount = new BigDecimal("250.50");
            PostTransactionRequest request = new PostTransactionRequest(
                    "Sale",
                    null,
                    List.of(
                            new EntryRequest(1L, amount, EntryType.DEBIT),
                            new EntryRequest(2L, amount, EntryType.CREDIT)
                    ));
            when(accountRepository.findByIdWithOptimisticLock(1L)).thenReturn(Optional.of(assetAccount));
            when(accountRepository.findByIdWithOptimisticLock(2L)).thenReturn(Optional.of(liabilityAccount));
            stubTransactionSaveWithIds();

            ledgerService.postTransaction(request);

            assertThat(assetAccount.getBalance()).isEqualByComparingTo("250.50");
            assertThat(liabilityAccount.getBalance()).isEqualByComparingTo("250.50");
        }

        @Test
        void postTransaction_multipleEntries_sumsDebitsAndCreditsCorrectly() {
            PostTransactionRequest request = new PostTransactionRequest(
                    "Split",
                    null,
                    List.of(
                            new EntryRequest(1L, new BigDecimal("60"), EntryType.DEBIT),
                            new EntryRequest(1L, new BigDecimal("40"), EntryType.DEBIT),
                            new EntryRequest(2L, new BigDecimal("100"), EntryType.CREDIT)
                    ));
            when(accountRepository.findByIdWithOptimisticLock(1L)).thenReturn(Optional.of(assetAccount));
            when(accountRepository.findByIdWithOptimisticLock(2L)).thenReturn(Optional.of(liabilityAccount));
            stubTransactionSaveWithIds();

            ledgerService.postTransaction(request);

            assertThat(assetAccount.getBalance()).isEqualByComparingTo("100");
            assertThat(liabilityAccount.getBalance()).isEqualByComparingTo("100");
        }
    }

    private void stubTransactionSaveWithIds() {
        doAnswer((Answer<Transaction>) invocation -> {
            Transaction tx = invocation.getArgument(0);
            setTransactionId(tx, 100L);
            int i = 0;
            for (Entry e : tx.getEntries()) {
                setEntryId(e, (long) (200 + i++));
            }
            return tx;
        }).when(transactionRepository).save(any(Transaction.class));
    }

    private static void setAccountId(Account account, Long id) {
        try {
            var idField = Account.class.getDeclaredField("id");
            idField.setAccessible(true);
            idField.set(account, id);
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    private static void setTransactionId(Transaction transaction, Long id) {
        try {
            var idField = Transaction.class.getDeclaredField("id");
            idField.setAccessible(true);
            idField.set(transaction, id);
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    private static void setEntryId(Entry entry, Long id) {
        try {
            var idField = Entry.class.getDeclaredField("id");
            idField.setAccessible(true);
            idField.set(entry, id);
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }
}

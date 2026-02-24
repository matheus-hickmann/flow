package com.flow.ledger.service;

import com.flow.ledger.dto.AccountResponse;
import com.flow.ledger.dto.AdjustBalanceRequest;
import com.flow.ledger.dto.CreateAccountRequest;
import com.flow.ledger.exception.AccountNotFoundException;
import com.flow.ledger.model.entity.Account;
import com.flow.ledger.model.entity.AccountType;
import com.flow.ledger.repository.AccountRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.math.BigDecimal;
import java.util.List;
import java.util.Optional;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class AccountServiceTest {

    @Mock
    private AccountRepository accountRepository;

    @InjectMocks
    private AccountService accountService;

    private Account savedAccount;

    @BeforeEach
    void setUp() {
        savedAccount = new Account("ITAU", "Itaú", AccountType.ASSET, "#0066cc");
        savedAccount.setBalance(new BigDecimal("1000.00"));
    }

    @Nested
    @DisplayName("listAll")
    class ListAll {

        @Test
        void accountService_listAll_returnsAllAccounts() {
            when(accountRepository.findAll()).thenReturn(List.of(savedAccount));

            List<AccountResponse> result = accountService.listAll();

            assertThat(result).hasSize(1);
            assertThat(result.get(0).id()).isEqualTo(savedAccount.getId());
            assertThat(result.get(0).name()).isEqualTo("Itaú");
            assertThat(result.get(0).balance()).isEqualByComparingTo(new BigDecimal("1000.00"));
            assertThat(result.get(0).color()).isEqualTo("#0066cc");
        }

        @Test
        void accountService_listAll_whenEmpty_returnsEmptyList() {
            when(accountRepository.findAll()).thenReturn(List.of());

            List<AccountResponse> result = accountService.listAll();

            assertThat(result).isEmpty();
        }
    }

    @Nested
    @DisplayName("create")
    class Create {

        @Test
        void accountService_create_persistsAccountWithInitialBalanceAndColor() {
            when(accountRepository.findByCode("NUBANK")).thenReturn(Optional.empty());
            when(accountRepository.save(org.mockito.ArgumentMatchers.any(Account.class))).thenAnswer(inv -> {
                Account a = inv.getArgument(0);
                return a;
            });

            CreateAccountRequest request = new CreateAccountRequest(
                    "Nubank",
                    new BigDecimal("500.00"),
                    "#820ad1"
            );
            AccountResponse result = accountService.create(request);

            ArgumentCaptor<Account> captor = ArgumentCaptor.forClass(Account.class);
            verify(accountRepository).save(captor.capture());
            Account saved = captor.getValue();
            assertThat(saved.getCode()).isEqualTo("NUBANK");
            assertThat(saved.getName()).isEqualTo("Nubank");
            assertThat(saved.getType()).isEqualTo(AccountType.ASSET);
            assertThat(saved.getBalance()).isEqualByComparingTo(new BigDecimal("500.00"));
            assertThat(saved.getColor()).isEqualTo("#820ad1");
            assertThat(result.name()).isEqualTo("Nubank");
            assertThat(result.balance()).isEqualByComparingTo(new BigDecimal("500.00"));
        }
    }

    @Nested
    @DisplayName("adjustBalance")
    class AdjustBalance {

        @Test
        void accountService_adjustBalance_updatesBalance() {
            when(accountRepository.findByIdWithOptimisticLock(1L)).thenReturn(Optional.of(savedAccount));
            when(accountRepository.save(org.mockito.ArgumentMatchers.any(Account.class))).thenAnswer(inv -> inv.getArgument(0));

            AccountResponse result = accountService.adjustBalance(1L, new AdjustBalanceRequest(new BigDecimal("2500.50")));

            verify(accountRepository).save(savedAccount);
            assertThat(savedAccount.getBalance()).isEqualByComparingTo(new BigDecimal("2500.50"));
            assertThat(result.balance()).isEqualByComparingTo(new BigDecimal("2500.50"));
        }

        @Test
        void accountService_adjustBalance_whenAccountNotFound_throws() {
            when(accountRepository.findByIdWithOptimisticLock(999L)).thenReturn(Optional.empty());

            assertThatThrownBy(() -> accountService.adjustBalance(999L, new AdjustBalanceRequest(BigDecimal.ZERO)))
                    .isInstanceOf(AccountNotFoundException.class)
                    .hasMessageContaining("999");
        }
    }
}

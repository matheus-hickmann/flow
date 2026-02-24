package com.flow.ledger.service;

import com.flow.ledger.dto.AccountResponse;
import com.flow.ledger.dto.AdjustBalanceRequest;
import com.flow.ledger.dto.CreateAccountRequest;
import com.flow.ledger.exception.AccountNotFoundException;
import com.flow.ledger.model.entity.Account;
import com.flow.ledger.model.entity.AccountType;
import com.flow.ledger.repository.AccountRepository;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.text.Normalizer;
import java.util.List;
import java.util.Optional;
import java.util.regex.Pattern;

@Service
public class AccountService {

    private static final int CODE_MAX_LENGTH = 50;
    private static final Pattern NON_ALPHANUMERIC = Pattern.compile("[^A-Z0-9]");

    private final AccountRepository accountRepository;

    public AccountService(AccountRepository accountRepository) {
        this.accountRepository = accountRepository;
    }

    public List<AccountResponse> listAll() {
        return accountRepository.findAll().stream()
                .map(AccountService::toResponse)
                .toList();
    }

    @Transactional
    public AccountResponse create(CreateAccountRequest request) {
        String code = generateUniqueCode(request.name());
        Account account = new Account(code, request.name(), AccountType.ASSET, request.colorOrDefault());
        account.setBalance(request.initialBalance());
        account = accountRepository.save(account);
        return toResponse(account);
    }

    @Transactional
    public AccountResponse adjustBalance(Long id, AdjustBalanceRequest request) {
        Account account = accountRepository.findByIdWithOptimisticLock(id)
                .orElseThrow(() -> new AccountNotFoundException(id));
        account.setBalance(request.newBalance());
        account = accountRepository.save(account);
        return toResponse(account);
    }

    private String generateUniqueCode(String name) {
        String base = normalizeToCode(name);
        if (base.length() > CODE_MAX_LENGTH) {
            base = base.substring(0, CODE_MAX_LENGTH);
        }
        if (base.isEmpty()) {
            base = "ACC";
        }
        String code = base;
        int suffix = 1;
        while (accountRepository.findByCode(code).isPresent()) {
            code = base + "_" + (suffix++);
            if (code.length() > CODE_MAX_LENGTH) {
                code = base.substring(0, CODE_MAX_LENGTH - String.valueOf(suffix).length() - 1) + "_" + suffix;
            }
        }
        return code;
    }

    private static String normalizeToCode(String name) {
        String nfd = Normalizer.normalize(name.trim(), Normalizer.Form.NFD);
        String ascii = nfd.replaceAll("\\p{M}", "");
        return NON_ALPHANUMERIC.matcher(ascii.toUpperCase()).replaceAll("").replaceAll("_+", "_");
    }

    private static AccountResponse toResponse(Account account) {
        return new AccountResponse(
                account.getId(),
                account.getCode(),
                account.getName(),
                account.getType(),
                account.getBalance(),
                account.getColor()
        );
    }
}

package com.flow.ledger.exception;

public class InvalidLedgerTransactionException extends RuntimeException {

    public InvalidLedgerTransactionException(String message) {
        super(message);
    }
}

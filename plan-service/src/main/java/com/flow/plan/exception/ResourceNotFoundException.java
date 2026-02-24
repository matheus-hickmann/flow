package com.flow.plan.exception;

public class ResourceNotFoundException extends RuntimeException {

    public ResourceNotFoundException(String resourceName, Long id) {
        super(id != null
                ? resourceName + " not found with id: " + id
                : resourceName + " not found");
    }
}

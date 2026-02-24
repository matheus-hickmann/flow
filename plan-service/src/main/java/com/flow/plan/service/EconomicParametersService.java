package com.flow.plan.service;

import com.flow.plan.dto.EconomicParametersRequest;
import com.flow.plan.dto.EconomicParametersResponse;
import com.flow.plan.exception.ResourceNotFoundException;
import com.flow.plan.model.entity.EconomicParameters;
import com.flow.plan.repository.EconomicParametersRepository;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service
public class EconomicParametersService {

    private final EconomicParametersRepository repository;

    public EconomicParametersService(EconomicParametersRepository repository) {
        this.repository = repository;
    }

    public EconomicParametersResponse get() {
        EconomicParameters entity = repository.findFirstByOrderByIdAsc()
                .orElseThrow(() -> new ResourceNotFoundException("EconomicParameters", null));
        return toResponse(entity);
    }

    @Transactional
    public EconomicParametersResponse createOrUpdate(EconomicParametersRequest request) {
        EconomicParameters entity = repository.findFirstByOrderByIdAsc()
                .orElseGet(() -> {
                    EconomicParameters newEntity = new EconomicParameters(request.selicRate(), request.ipcaRate());
                    return repository.save(newEntity);
                });
        entity.setSelicRate(request.selicRate());
        entity.setIpcaRate(request.ipcaRate());
        EconomicParameters saved = repository.save(entity);
        return toResponse(saved);
    }

    private static EconomicParametersResponse toResponse(EconomicParameters entity) {
        return new EconomicParametersResponse(
                entity.getId(),
                entity.getSelicRate(),
                entity.getIpcaRate(),
                entity.getUpdatedAt()
        );
    }
}

package com.flow.plan.repository;

import com.flow.plan.model.entity.EconomicParameters;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.Optional;

public interface EconomicParametersRepository extends JpaRepository<EconomicParameters, Long> {

    Optional<EconomicParameters> findFirstByOrderByIdAsc();
}

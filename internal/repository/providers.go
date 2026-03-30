package repository

import (
	domain "envmn/internal/domain/environment/services"
	"envmn/internal/repository/postgres"
	pgqueries "envmn/internal/repository/queries/postgres"
	"envmn/internal/service/ports"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ProvideEnvironmentRepository(db *pgxpool.Pool) ports.EnvironmentRepository {
	return postgres.NewEnvironmentsRepository(pgqueries.New(db))
}

func ProvideEnvironmentPoliciesRepository(db *pgxpool.Pool) ports.EnvironmentPoliciesRepository {
	return postgres.NewAccessPoliciesRepository(pgqueries.New(db))
}

func ProvideAccessPolicyRepository(db *pgxpool.Pool) domain.AccessPolicyFinderSaver {
	return postgres.NewAccessPoliciesRepository(pgqueries.New(db))
}

func ProvideAccessPolicyFinderSaver(db *pgxpool.Pool) ports.AccessPolicyRepository {
	return postgres.NewAccessPoliciesRepository(pgqueries.New(db))
}

func ProvideEnvironmentVariablesRepository(db *pgxpool.Pool) ports.EnvironmentVariablesRepository {
	return postgres.NewVariablesRepository(pgqueries.New(db))
}

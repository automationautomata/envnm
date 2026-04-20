package postgres

import (
	"context"
	"envmn/internal/domain/environment/entities"
	"envmn/internal/repository/postgres/queries"

	"github.com/google/uuid"
)

type accessPoliciesRepository struct {
	conn *connection
}

func NewAccessPoliciesRepository(conn *connection) *accessPoliciesRepository {
	return &accessPoliciesRepository{conn: conn}
}

func (r *accessPoliciesRepository) Save(ctx context.Context, policy *entities.AccessPolicy) error {
	err := r.conn.CreatePolicy(ctx, queries.CreatePolicyParams{
		ID:   policy.ID,
		Name: policy.Name,
		Key:  policy.Key,
	})

	if err != nil {
		return err
	}
	return nil
}

func (r *accessPoliciesRepository) FindByKey(ctx context.Context, key string) (*entities.AccessPolicy, error) {
	row, err := r.conn.FindPolicyByKey(ctx, key)
	if err != nil {
		return nil, err
	}

	return &entities.AccessPolicy{
		ID:   row.ID,
		Name: row.Name,
		Key:  row.Key,
	}, nil
}

func (r *accessPoliciesRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.AccessPolicy, error) {
	row, err := r.conn.FindPolicyByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &entities.AccessPolicy{
		ID:   row.ID,
		Name: row.Name,
		Key:  row.Key,
	}, nil
}

func (r *accessPoliciesRepository) FindByName(ctx context.Context, name string) (*entities.AccessPolicy, error) {
	row, err := r.conn.FindPolicyByName(ctx, name)
	if err != nil {
		return nil, err
	}

	return &entities.AccessPolicy{
		ID:   row.ID,
		Name: row.Name,
		Key:  row.Key,
	}, nil
}

func (r *accessPoliciesRepository) GetEnvironmentPolicies(ctx context.Context, envID uuid.UUID) ([]*entities.AccessPolicy, error) {
	rows, err := r.conn.GetPoliciesByEnv(ctx, envID)
	if err != nil {
		return nil, err
	}

	res := make([]*entities.AccessPolicy, len(rows))

	for i, row := range rows {
		res[i] = &entities.AccessPolicy{
			ID:   row.ID,
			Name: row.Name,
			Key:  row.Key,
		}
	}

	return res, nil
}

func (r *accessPoliciesRepository) AddToEnvironment(ctx context.Context, envID uuid.UUID, policy *entities.AccessPolicy) error {
	err := r.conn.CreatePolicy(ctx, queries.CreatePolicyParams{
		ID:   policy.ID,
		Name: policy.Name,
		Key:  policy.Key,
	})
	if err != nil {
		return err
	}

	return r.conn.AddPolicyToEnvironment(ctx, queries.AddPolicyToEnvironmentParams{
		EnvironmentID:  envID,
		AccessPolicyID: policy.ID,
		ChangesAllowed: true, // или из policy если есть
	})
}

func (r *accessPoliciesRepository) DeleteFromEnvironment(ctx context.Context, envID uuid.UUID, policyID uuid.UUID) error {
	return r.conn.DeletePolicyFromEnvironment(ctx, queries.DeletePolicyFromEnvironmentParams{
		EnvironmentID:  envID,
		AccessPolicyID: policyID,
	})
}

func (r *accessPoliciesRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.conn.DeletePolicy(ctx, id)
}

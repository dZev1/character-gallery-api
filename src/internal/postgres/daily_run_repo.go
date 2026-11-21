package postgres

import (
	"context"
	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/daily_run"
	"dZev1/character-gallery/internal/enemies"
	"dZev1/character-gallery/internal/postgres/db"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ daily_run.Repository = (*dailyRunRepo)(nil)

type dailyRunRepo struct {
	pool *pgxpool.Pool
}

func NewDailyRunRepo(pool *pgxpool.Pool) *dailyRunRepo {
	return &dailyRunRepo{pool: pool}
}

func (r *dailyRunRepo) q(exec db.DBTX) *db.Queries {
	return db.New(exec)
}

func (r *dailyRunRepo) CreateDailyRun(ctx context.Context, exec db.DBTX, run *daily_run.DailyRun) (*daily_run.DailyRun, error) {
	q := r.q(exec)

	visited := run.VisitedNodes
	if visited == nil {
		visited = []string{}
	}

	row, err := q.CreateDailyRun(ctx, db.CreateDailyRunParams{
		RunDate:        toPGDate(run.RunDate),
		Tier:           int16(run.Tier),
		UserID:         run.UserID,
		CharacterID:    int64(run.CharacterID),
		TotalFloors:    int16(run.TotalFloors),
		CurrentFloor:   int16(run.CurrentFloor),
		CurrentNodeID:  run.CurrentNodeID,
		VisitedNodes:   visited,
		TurnsTaken:     int32(run.TurnsTaken),
		Score:          int32(run.Score),
		Status:         string(run.Status),
		CurrentEnemyID: toEnemyInt8(run.CurrentEnemyID),
		CurrentEnemyHp: toUintInt4(run.CurrentEnemyHP),
	})
	if err != nil {
		return nil, fmt.Errorf("create daily run: %w", err)
	}

	return toDomainDailyRun(row), nil
}

func (r *dailyRunRepo) GetDailyRun(ctx context.Context, exec db.DBTX, runDate time.Time, tier uint8, characterID characters.ID) (*daily_run.DailyRun, error) {
	q := r.q(exec)

	row, err := q.GetDailyRun(ctx, db.GetDailyRunParams{
		RunDate:     toPGDate(runDate),
		Tier:        int16(tier),
		CharacterID: int64(characterID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("daily run for character %d: %w", characterID, daily_run.ErrNotFound)
		}
		return nil, fmt.Errorf("get daily run: %w", err)
	}

	return toDomainDailyRun(row), nil
}

func (r *dailyRunRepo) GetDailyRunByID(ctx context.Context, exec db.DBTX, id daily_run.ID) (*daily_run.DailyRun, error) {
	q := r.q(exec)

	row, err := q.GetDailyRunByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("daily run %d: %w", id, daily_run.ErrNotFound)
		}
		return nil, fmt.Errorf("get daily run %d: %w", id, err)
	}

	return toDomainDailyRun(row), nil
}

func (r *dailyRunRepo) GetActiveDailyRunForCharacter(ctx context.Context, exec db.DBTX, runDate time.Time, characterID characters.ID) (*daily_run.DailyRun, error) {
	q := r.q(exec)

	row, err := q.GetActiveDailyRunForCharacter(ctx, db.GetActiveDailyRunForCharacterParams{
		RunDate:     toPGDate(runDate),
		CharacterID: int64(characterID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("active daily run for character %d: %w", characterID, daily_run.ErrNotFound)
		}
		return nil, fmt.Errorf("get active daily run for character %d: %w", characterID, err)
	}

	return toDomainDailyRun(row), nil
}

func (r *dailyRunRepo) ListActiveDailyRunsByUser(ctx context.Context, exec db.DBTX, runDate time.Time, userID uuid.UUID) ([]daily_run.DailyRun, error) {
	q := r.q(exec)

	rows, err := q.ListActiveDailyRunsByUser(ctx, db.ListActiveDailyRunsByUserParams{
		RunDate: toPGDate(runDate),
		UserID:  userID,
	})
	if err != nil {
		return nil, fmt.Errorf("list active daily runs for user %s: %w", userID, err)
	}

	result := make([]daily_run.DailyRun, len(rows))
	for i, row := range rows {
		result[i] = *toDomainDailyRun(row)
	}
	return result, nil
}

func (r *dailyRunRepo) UpdateDailyRunState(ctx context.Context, exec db.DBTX, run *daily_run.DailyRun) (*daily_run.DailyRun, error) {
	q := r.q(exec)

	visited := run.VisitedNodes
	if visited == nil {
		visited = []string{}
	}

	row, err := q.UpdateDailyRunState(ctx, db.UpdateDailyRunStateParams{
		ID:             int64(run.ID),
		CurrentFloor:   int16(run.CurrentFloor),
		CurrentNodeID:  run.CurrentNodeID,
		VisitedNodes:   visited,
		TurnsTaken:     int32(run.TurnsTaken),
		Score:          int32(run.Score),
		Status:         string(run.Status),
		CurrentEnemyID: toEnemyInt8(run.CurrentEnemyID),
		CurrentEnemyHp: toUintInt4(run.CurrentEnemyHP),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("update daily run %d: %w", run.ID, daily_run.ErrNotFound)
		}
		return nil, fmt.Errorf("update daily run %d: %w", run.ID, err)
	}

	return toDomainDailyRun(row), nil
}

func (r *dailyRunRepo) UpdateDailyRunEnemyHP(ctx context.Context, exec db.DBTX, id daily_run.ID, enemyHP *uint, turnsTaken uint) (*daily_run.DailyRun, error) {
	q := r.q(exec)

	row, err := q.UpdateDailyRunEnemyHP(ctx, db.UpdateDailyRunEnemyHPParams{
		ID:             int64(id),
		CurrentEnemyHp: toUintInt4(enemyHP),
		TurnsTaken:     int32(turnsTaken),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("update daily run enemy hp %d: %w", id, daily_run.ErrNotFound)
		}
		return nil, fmt.Errorf("update daily run enemy hp %d: %w", id, err)
	}

	return toDomainDailyRun(row), nil
}

func (r *dailyRunRepo) UpsertDailyScore(ctx context.Context, exec db.DBTX, score *daily_run.DailyScore) (*daily_run.DailyScore, error) {
	q := r.q(exec)

	row, err := q.UpsertDailyScore(ctx, db.UpsertDailyScoreParams{
		RunDate:           toPGDate(score.RunDate),
		Tier:              int16(score.Tier),
		UserID:            score.UserID,
		CharacterID:       int64(score.CharacterID),
		CharacterName:     score.CharacterName,
		CharacterClass:    string(score.CharacterClass),
		CharacterLevel:    int32(score.CharacterLevel),
		CharacterPrestige: int32(score.CharacterPrestige),
		Score:             int32(score.Score),
		FloorsCleared:     int16(score.FloorsCleared),
		TotalFloors:       int16(score.TotalFloors),
		TurnsTaken:        int32(score.TurnsTaken),
		HpRemaining:       int32(score.HpRemaining),
		Completed:         score.Completed,
	})
	if err != nil {
		return nil, fmt.Errorf("upsert daily score: %w", err)
	}

	return toDomainDailyScore(row), nil
}

func (r *dailyRunRepo) GetDailyLeaderboard(ctx context.Context, exec db.DBTX, runDate time.Time, tier uint8, limit, offset int) ([]daily_run.DailyScore, uint64, error) {
	q := r.q(exec)

	pgDate := toPGDate(runDate)
	rows, err := q.GetDailyLeaderboard(ctx, db.GetDailyLeaderboardParams{
		RunDate: pgDate,
		Tier:    int16(tier),
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("get daily leaderboard: %w", err)
	}

	total, err := q.CountDailyLeaderboard(ctx, db.CountDailyLeaderboardParams{
		RunDate: pgDate,
		Tier:    int16(tier),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("count daily leaderboard: %w", err)
	}

	result := make([]daily_run.DailyScore, len(rows))
	for i, row := range rows {
		result[i] = daily_run.DailyScore{
			ID:                uint64(row.ID),
			RunDate:           row.RunDate.Time,
			Tier:              uint8(row.Tier),
			UserID:            row.UserID,
			Username:          row.Username,
			CharacterID:       characters.ID(row.CharacterID),
			CharacterName:     row.CharacterName,
			CharacterClass:    characters.Class(row.CharacterClass),
			CharacterLevel:    uint8(row.CharacterLevel),
			CharacterPrestige: uint(row.CharacterPrestige),
			Score:             uint(row.Score),
			FloorsCleared:     uint8(row.FloorsCleared),
			TotalFloors:       uint8(row.TotalFloors),
			TurnsTaken:        uint(row.TurnsTaken),
			HpRemaining:       uint8(row.HpRemaining),
			Completed:         row.Completed,
			CreatedAt:         row.CreatedAt.Time,
		}
	}

	return result, uint64(total), nil
}

func (r *dailyRunRepo) GetUserDailyScores(ctx context.Context, exec db.DBTX, runDate time.Time, userID uuid.UUID) ([]daily_run.DailyScore, error) {
	q := r.q(exec)

	rows, err := q.GetUserDailyScores(ctx, db.GetUserDailyScoresParams{
		RunDate: toPGDate(runDate),
		UserID:  userID,
	})
	if err != nil {
		return nil, fmt.Errorf("get user daily scores: %w", err)
	}

	result := make([]daily_run.DailyScore, len(rows))
	for i, row := range rows {
		result[i] = *toDomainDailyScore(row)
	}
	return result, nil
}

func toDomainDailyRun(row *db.DailyRun) *daily_run.DailyRun {
	return &daily_run.DailyRun{
		ID:             daily_run.ID(row.ID),
		RunDate:        row.RunDate.Time,
		Tier:           uint8(row.Tier),
		UserID:         row.UserID,
		CharacterID:    characters.ID(row.CharacterID),
		TotalFloors:    uint8(row.TotalFloors),
		CurrentFloor:   uint8(row.CurrentFloor),
		CurrentNodeID:  row.CurrentNodeID,
		VisitedNodes:   row.VisitedNodes,
		TurnsTaken:     uint(row.TurnsTaken),
		Score:          uint(row.Score),
		Status:         daily_run.Status(row.Status),
		CurrentEnemyID: fromEnemyInt8(row.CurrentEnemyID),
		CurrentEnemyHP: fromUintInt4(row.CurrentEnemyHp),
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
	}
}

func toDomainDailyScore(row *db.DailyScore) *daily_run.DailyScore {
	return &daily_run.DailyScore{
		ID:                uint64(row.ID),
		RunDate:           row.RunDate.Time,
		Tier:              uint8(row.Tier),
		UserID:            row.UserID,
		CharacterID:       characters.ID(row.CharacterID),
		CharacterName:     row.CharacterName,
		CharacterClass:    characters.Class(row.CharacterClass),
		CharacterLevel:    uint8(row.CharacterLevel),
		CharacterPrestige: uint(row.CharacterPrestige),
		Score:             uint(row.Score),
		FloorsCleared:     uint8(row.FloorsCleared),
		TotalFloors:       uint8(row.TotalFloors),
		TurnsTaken:        uint(row.TurnsTaken),
		HpRemaining:       uint8(row.HpRemaining),
		Completed:         row.Completed,
		CreatedAt:         row.CreatedAt.Time,
	}
}

func toPGDate(t time.Time) pgtype.Date {
	y, m, d := t.UTC().Date()
	return pgtype.Date{
		Time:  time.Date(y, m, d, 0, 0, 0, 0, time.UTC),
		Valid: true,
	}
}

func toEnemyInt8(ptr *enemies.ID) pgtype.Int8 {
	if ptr != nil {
		return pgtype.Int8{Int64: int64(*ptr), Valid: true}
	}
	return pgtype.Int8{}
}

func fromEnemyInt8(f pgtype.Int8) *enemies.ID {
	if f.Valid {
		v := enemies.ID(f.Int64)
		return &v
	}
	return nil
}

func toUintInt4(ptr *uint) pgtype.Int4 {
	if ptr != nil {
		return pgtype.Int4{Int32: int32(*ptr), Valid: true}
	}
	return pgtype.Int4{}
}

func fromUintInt4(f pgtype.Int4) *uint {
	if f.Valid {
		v := uint(f.Int32)
		return &v
	}
	return nil
}

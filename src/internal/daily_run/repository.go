package daily_run

import (
	"context"
	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/postgres/db"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	CreateDailyRun(ctx context.Context, exec db.DBTX, run *DailyRun) (*DailyRun, error)
	GetDailyRun(ctx context.Context, exec db.DBTX, runDate time.Time, tier uint8, characterID characters.ID) (*DailyRun, error)
	GetDailyRunByID(ctx context.Context, exec db.DBTX, id ID) (*DailyRun, error)
	GetActiveDailyRunForCharacter(ctx context.Context, exec db.DBTX, runDate time.Time, characterID characters.ID) (*DailyRun, error)
	ListActiveDailyRunsByUser(ctx context.Context, exec db.DBTX, runDate time.Time, userID uuid.UUID) ([]DailyRun, error)
	UpdateDailyRunState(ctx context.Context, exec db.DBTX, run *DailyRun) (*DailyRun, error)
	UpdateDailyRunEnemyHP(ctx context.Context, exec db.DBTX, id ID, enemyHP *uint, turnsTaken uint) (*DailyRun, error)

	UpsertDailyScore(ctx context.Context, exec db.DBTX, score *DailyScore) (*DailyScore, error)
	GetDailyLeaderboard(ctx context.Context, exec db.DBTX, runDate time.Time, tier uint8, limit, offset int) ([]DailyScore, uint64, error)
	GetUserDailyScores(ctx context.Context, exec db.DBTX, runDate time.Time, userID uuid.UUID) ([]DailyScore, error)
}

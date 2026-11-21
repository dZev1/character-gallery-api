package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/daily_run"
	"dZev1/character-gallery/internal/enemies"
)

func TestDailyRunRepo_CreateGetAndPauseResume(t *testing.T) {
	tt := newTestTx(t)
	repo := newDailyRunRepo(t)
	enemyRepository := newEnemyRepo(t)
	ctx := context.Background()

	char := createTestCharacter(t, tt.Q)
	enemy, err := enemyRepository.SaveEnemy(ctx, tt.Tx, &enemies.Enemy{
		Name: "Cave Troll", Type: enemies.Elite, Tier: 3,
		Description: "Regenerating brute", HpMax: 90, ArmorClass: 15,
		Damage: 12, XPReward: 150, ScoreReward: 120,
	})
	if err != nil {
		t.Fatal(err)
	}

	runDate := time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC)

	created, err := repo.CreateDailyRun(ctx, tt.Tx, &daily_run.DailyRun{
		RunDate:       runDate,
		Tier:          3,
		UserID:        char.OwnerID,
		CharacterID:   characters.ID(char.ID),
		TotalFloors:   20,
		CurrentFloor:  1,
		CurrentNodeID: "f1_n1",
		VisitedNodes:  []string{"f1_n1"},
		TurnsTaken:    0,
		Score:         0,
		Status:        daily_run.InProgress,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero daily run ID")
	}

	// Simulate entering combat on floor 5 and pausing mid-fight ("going to eat").
	enemyID := enemy.ID
	enemyHP := uint(45)
	created.CurrentFloor = 5
	created.CurrentNodeID = "f5_n2"
	created.VisitedNodes = []string{"f1_n1", "f2_n1", "f3_n2", "f4_n1", "f5_n2"}
	created.TurnsTaken = 11
	created.Score = 350
	created.Status = daily_run.InCombat
	created.CurrentEnemyID = &enemyID
	created.CurrentEnemyHP = &enemyHP

	updated, err := repo.UpdateDailyRunState(ctx, tt.Tx, created)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != daily_run.InCombat {
		t.Fatalf("got status %q, want %q", updated.Status, daily_run.InCombat)
	}
	if updated.CurrentEnemyID == nil || *updated.CurrentEnemyID != enemy.ID {
		t.Fatalf("got current_enemy_id %v, want %d", updated.CurrentEnemyID, enemy.ID)
	}
	if updated.CurrentEnemyHP == nil || *updated.CurrentEnemyHP != 45 {
		t.Fatalf("got current_enemy_hp %v, want 45", updated.CurrentEnemyHP)
	}

	// Simulate updating enemy HP after one more combat turn.
	nextHP := uint(20)
	afterTurn, err := repo.UpdateDailyRunEnemyHP(ctx, tt.Tx, created.ID, &nextHP, 12)
	if err != nil {
		t.Fatal(err)
	}
	if afterTurn.CurrentEnemyHP == nil || *afterTurn.CurrentEnemyHP != 20 || afterTurn.TurnsTaken != 12 {
		t.Fatalf("unexpected state after turn: hp=%v turns=%d", afterTurn.CurrentEnemyHP, afterTurn.TurnsTaken)
	}

	// Resume active run for character.
	resumed, err := repo.GetActiveDailyRunForCharacter(ctx, tt.Tx, runDate, characters.ID(char.ID))
	if err != nil {
		t.Fatal(err)
	}
	if resumed.ID != created.ID || len(resumed.VisitedNodes) != 5 {
		t.Fatalf("unexpected resumed run: %+v", resumed)
	}

	activeList, err := repo.ListActiveDailyRunsByUser(ctx, tt.Tx, runDate, char.OwnerID)
	if err != nil {
		t.Fatal(err)
	}
	if len(activeList) != 1 {
		t.Fatalf("expected 1 active run for user, got %d", len(activeList))
	}

	// Finish the run with Victory.
	resumed.Status = daily_run.Victory
	resumed.CurrentFloor = 20
	resumed.Score = 2450
	resumed.CurrentEnemyID = nil
	resumed.CurrentEnemyHP = nil

	finished, err := repo.UpdateDailyRunState(ctx, tt.Tx, resumed)
	if err != nil {
		t.Fatal(err)
	}
	if finished.Status != daily_run.Victory || finished.CurrentEnemyID != nil {
		t.Fatalf("expected completed run without active enemy, got %+v", finished)
	}

	// Completed runs should no longer be returned by GetActiveDailyRunForCharacter.
	if _, err := repo.GetActiveDailyRunForCharacter(ctx, tt.Tx, runDate, characters.ID(char.ID)); !errors.Is(err, daily_run.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for active run after victory, got %v", err)
	}
}

func TestDailyRunRepo_LeaderboardAndScores(t *testing.T) {
	tt := newTestTx(t)
	repo := newDailyRunRepo(t)
	ctx := context.Background()

	char1 := createTestCharacter(t, tt.Q)
	char2 := createTestCharacter(t, tt.Q)
	runDate := time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC)

	_, err := repo.UpsertDailyScore(ctx, tt.Tx, &daily_run.DailyScore{
		RunDate:           runDate,
		Tier:              1,
		UserID:            char1.OwnerID,
		CharacterID:       characters.ID(char1.ID),
		CharacterName:     "Aragorn",
		CharacterClass:    characters.Ranger,
		CharacterLevel:    2,
		CharacterPrestige: 0,
		Score:             1200,
		FloorsCleared:     10,
		TotalFloors:       10,
		TurnsTaken:        25,
		HpRemaining:       14,
		Completed:         true,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = repo.UpsertDailyScore(ctx, tt.Tx, &daily_run.DailyScore{
		RunDate:           runDate,
		Tier:              1,
		UserID:            char2.OwnerID,
		CharacterID:       characters.ID(char2.ID),
		CharacterName:     "Gandalf",
		CharacterClass:    characters.Wizard,
		CharacterLevel:    2,
		CharacterPrestige: 1,
		Score:             1550,
		FloorsCleared:     10,
		TotalFloors:       10,
		TurnsTaken:        19,
		HpRemaining:       20,
		Completed:         true,
	})
	if err != nil {
		t.Fatal(err)
	}

	board, total, err := repo.GetDailyLeaderboard(ctx, tt.Tx, runDate, 1, 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(board) != 2 {
		t.Fatalf("expected 2 leaderboard entries, got total=%d len=%d", total, len(board))
	}
	if board[0].CharacterName != "Gandalf" || board[0].Score != 1550 {
		t.Fatalf("expected Gandalf first with score 1550, got %+v", board[0])
	}
	if board[0].Username == "" {
		t.Errorf("expected joined username in leaderboard row")
	}

	userScores, err := repo.GetUserDailyScores(ctx, tt.Tx, runDate, char1.OwnerID)
	if err != nil {
		t.Fatal(err)
	}
	if len(userScores) != 1 || userScores[0].Score != 1200 {
		t.Fatalf("unexpected user daily scores: %+v", userScores)
	}
}

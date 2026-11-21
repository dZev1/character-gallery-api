package postgres

import (
	"context"
	"errors"
	"testing"

	"dZev1/character-gallery/internal/enemies"
)

func TestEnemyRepo_SaveAndFindEnemy(t *testing.T) {
	tt := newTestTx(t)
	repo := newEnemyRepo(t)
	ctx := context.Background()

	enemy := &enemies.Enemy{
		Name:        "Goblin Scout",
		Type:        enemies.Normal,
		Tier:        1,
		Description: "A sneaky goblin armed with a rusty dagger",
		HpMax:       18,
		ArmorClass:  11,
		Damage:      4,
		XPReward:    25,
		ScoreReward: 50,
	}

	saved, err := repo.SaveEnemy(ctx, tt.Tx, enemy)
	if err != nil {
		t.Fatal(err)
	}
	if saved.ID == 0 {
		t.Fatal("expected non-zero ID")
	}

	got, err := repo.FindEnemy(ctx, tt.Tx, saved.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Goblin Scout" {
		t.Errorf("got name %q, want %q", got.Name, "Goblin Scout")
	}
	if got.Type != enemies.Normal {
		t.Errorf("got type %q, want %q", got.Type, enemies.Normal)
	}
	if got.ArmorClass != 11 {
		t.Errorf("got ac %d, want 11", got.ArmorClass)
	}
	if got.ScoreReward != 50 {
		t.Errorf("got score_reward %d, want 50", got.ScoreReward)
	}
}

func TestEnemyRepo_FindEnemy_NotFound(t *testing.T) {
	tt := newTestTx(t)
	repo := newEnemyRepo(t)
	ctx := context.Background()

	_, err := repo.FindEnemy(ctx, tt.Tx, 99999)
	if err == nil || !errors.Is(err, enemies.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestEnemyRepo_FindByTierAndType(t *testing.T) {
	tt := newTestTx(t)
	repo := newEnemyRepo(t)
	ctx := context.Background()

	seeds := []enemies.Enemy{
		{
			Name: "Giant Rat", Type: enemies.Normal, Tier: 1,
			Description: "A oversized sewer rat", HpMax: 12, ArmorClass: 9,
			Damage: 3, XPReward: 15, ScoreReward: 30,
		},
		{
			Name: "Hobgoblin Captain", Type: enemies.Elite, Tier: 1,
			Description: "A disciplined goblin leader", HpMax: 35, ArmorClass: 13,
			Damage: 7, XPReward: 60, ScoreReward: 120,
		},
		{
			Name: "Ancient Red Dragon", Type: enemies.Boss, Tier: 5,
			Description: "Lord of flame and ash", HpMax: 450, ArmorClass: 24,
			Damage: 35, XPReward: 1000, ScoreReward: 500,
		},
	}

	if err := repo.SeedEnemies(ctx, tt.Tx, seeds); err != nil {
		t.Fatal(err)
	}

	all, total, err := repo.FindAllEnemies(ctx, tt.Tx, 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 3 || len(all) != 3 {
		t.Fatalf("expected 3 enemies, got total=%d len=%d", total, len(all))
	}

	tier1, err := repo.FindEnemiesByTier(ctx, tt.Tx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(tier1) != 2 {
		t.Fatalf("expected 2 tier-1 enemies, got %d", len(tier1))
	}

	elites, err := repo.FindEnemiesByTierAndType(ctx, tt.Tx, 1, enemies.Elite)
	if err != nil {
		t.Fatal(err)
	}
	if len(elites) != 1 || elites[0].Name != "Hobgoblin Captain" {
		t.Fatalf("expected Hobgoblin Captain elite, got %+v", elites)
	}
}

func TestEnemyRepo_UpdateAndDeleteEnemy(t *testing.T) {
	tt := newTestTx(t)
	repo := newEnemyRepo(t)
	ctx := context.Background()

	enemy, err := repo.SaveEnemy(ctx, tt.Tx, &enemies.Enemy{
		Name: "Skeleton Warrior", Type: enemies.Normal, Tier: 2,
		Description: "Undead soldier", HpMax: 28, ArmorClass: 13,
		Damage: 6, XPReward: 40, ScoreReward: 60,
	})
	if err != nil {
		t.Fatal(err)
	}

	enemy.Name = "Armored Skeleton"
	enemy.ArmorClass = 15
	enemy.HpMax = 34

	updated, err := repo.UpdateEnemy(ctx, tt.Tx, enemy)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Armored Skeleton" || updated.ArmorClass != 15 || updated.HpMax != 34 {
		t.Fatalf("unexpected updated enemy: %+v", updated)
	}

	if err := repo.DeleteEnemy(ctx, tt.Tx, enemy.ID); err != nil {
		t.Fatal(err)
	}

	if _, err := repo.FindEnemy(ctx, tt.Tx, enemy.ID); !errors.Is(err, enemies.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}

	if err := repo.DeleteEnemy(ctx, tt.Tx, enemy.ID); !errors.Is(err, enemies.ErrNotFound) {
		t.Fatalf("expected ErrNotFound deleting already deleted enemy, got %v", err)
	}
}

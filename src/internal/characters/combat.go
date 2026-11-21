package characters

// Tier returns the character's current Daily Run tier (1..5) derived from level.
// Levels 1-2 -> Tier 1, 3-4 -> Tier 2, ..., 9-10 -> Tier 5.
func (char *Character) Tier() uint8 {
	if char.Level <= 1 {
		return 1
	}
	t := ((char.Level - 1) / 2) + 1
	if t > 5 {
		return 5
	}
	return t
}

// StatMod computes the RPG modifier from a 1..99 stat (0..9).
func StatMod(stat uint8) int {
	return int(stat) / 10
}

// PrimaryStatValue returns the character's primary combat stat based on class.
func (char *Character) PrimaryStatValue() uint8 {
	if char.Stats == nil {
		return 10
	}
	switch char.Class {
	case Barbarian, Fighter, Paladin:
		return char.Stats.Strength
	case Rogue, Ranger, Monk:
		return char.Stats.Dexterity
	case Wizard:
		return char.Stats.Intelligence
	case Cleric, Druid:
		return char.Stats.Wisdom
	case Bard, Sorcerer, Warlock:
		return char.Stats.Charisma
	default:
		return char.Stats.Strength
	}
}

// BaseHPPerLevel returns the hit-die base HP gained per level for a class.
func (char *Character) BaseHPPerLevel() int {
	switch char.Class {
	case Barbarian:
		return 12
	case Fighter, Paladin, Ranger:
		return 10
	case Bard, Cleric, Druid, Monk, Rogue, Warlock:
		return 8
	case Sorcerer, Wizard:
		return 6
	default:
		return 8
	}
}

// StartingHP calculates level-1 max HP (used on creation or prestige reset).
func (char *Character) StartingHP() uint8 {
	conMod := 0
	if char.Stats != nil {
		conMod = StatMod(char.Stats.Constitution)
	}
	hp := char.BaseHPPerLevel() + conMod*2 + int(char.Prestige)*2
	if hp < 10 {
		hp = 10
	}
	if hp > 250 {
		hp = 250
	}
	return uint8(hp)
}

// HPGainOnLevelUp calculates how much HpMax increases when leveling up.
func (char *Character) HPGainOnLevelUp() uint8 {
	conMod := 0
	if char.Stats != nil {
		conMod = StatMod(char.Stats.Constitution)
	}
	gain := (char.BaseHPPerLevel() / 2) + 1 + conMod
	if gain < 2 {
		gain = 2
	}
	return uint8(gain)
}

// ArmorClass calculates the character's total AC given total defense from equipped gear.
func (char *Character) ArmorClass(equippedDefense int) int {
	dexMod := 0
	if char.Stats != nil {
		dexMod = StatMod(char.Stats.Dexterity)
	}
	// Scale equipped defense so item pool values (e.g. 6..19) map cleanly to D&D AC range.
	ac := 10 + dexMod + (equippedDefense / 3)
	if ac > 30 {
		ac = 30
	}
	return ac
}

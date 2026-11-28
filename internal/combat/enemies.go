package combat

import (
	"encoding/json"
	"fmt"
	"time"
)

// MinTelegraph is the shortest wind-up an enemy attack may have before
// anything can hurt the critter (docs/PLAN.md §2.7).
const MinTelegraph = 600 * time.Millisecond

// The behaviour names an enemy definition can use. Behaviours are code;
// everything else about an enemy is data.
const (
	// BehaviourHopToward hops toward the critter.
	BehaviourHopToward = "hop_toward"
	// BehaviourCharge shakes, then charges in a straight line.
	BehaviourCharge = "charge"
	// BehaviourPuff stands still and puffs spores in four directions.
	BehaviourPuff = "puff"
)

// EnemyName is an enemy's display name in each supported language.
type EnemyName struct {
	EN string `json:"en"`
	SV string `json:"sv"`
}

// Drop is one thing an enemy may leave behind when defeated.
type Drop struct {
	// Item is the id of what drops, such as "heart" or "berry".
	Item string `json:"item"`
	// Chance is the probability of the drop, in (0, 1].
	Chance float64 `json:"chance"`
}

// EnemyOverrides holds the values the Little Explorer preset replaces. A nil
// field keeps the base value.
type EnemyOverrides struct {
	HP     *int     `json:"hp,omitempty"`
	Damage *int     `json:"damage,omitempty"`
	Speed  *float64 `json:"speed,omitempty"`
}

// EnemyDef is one kind of enemy as described in enemies.json.
type EnemyDef struct {
	ID     string    `json:"id"`
	Name   EnemyName `json:"name"`
	Sprite string    `json:"sprite"`
	// Behaviour names the state machine that drives the enemy, one of
	// BehaviourHopToward, BehaviourCharge or BehaviourPuff.
	Behaviour string `json:"behaviour"`
	// HP is how many points of damage defeat the enemy.
	HP int `json:"hp"`
	// Damage is how many petals one of its hits takes.
	Damage int `json:"damage"`
	// Speed is in world pixels per second: how fast the enemy moves, or for
	// puff how fast its spores travel.
	Speed float64 `json:"speed"`
	// TelegraphSeconds is how long the wind-up before each attack lasts; at
	// least MinTelegraph.
	TelegraphSeconds float64 `json:"telegraphSeconds"`
	Drops            []Drop  `json:"drops"`
	// LittleExplorer holds the Little Explorer preset's overrides, or nil.
	LittleExplorer *EnemyOverrides `json:"littleExplorer,omitempty"`
}

// Telegraph returns TelegraphSeconds as a duration.
func (d EnemyDef) Telegraph() time.Duration {
	return time.Duration(d.TelegraphSeconds * float64(time.Second))
}

// ParseEnemies parses an enemies.json file, a JSON array of enemy
// definitions, and checks each one: a missing or duplicate id, an unknown
// behaviour, a hp, damage or speed that is not positive (for puff, speed is
// the speed of its spores, so it must be positive too), a telegraph shorter
// than MinTelegraph, a drop with an empty item or a chance outside (0, 1],
// and a Little Explorer override that breaks the same rules are all errors
// naming the enemy. Every definition it accepts can be built by NewBehaviour.
func ParseEnemies(data []byte) ([]EnemyDef, error) {
	var defs []EnemyDef
	if err := json.Unmarshal(data, &defs); err != nil {
		return nil, fmt.Errorf("parse enemies: %w", err)
	}
	seen := make(map[string]bool, len(defs))
	for i, d := range defs {
		if d.ID == "" {
			return nil, fmt.Errorf("enemy %d: missing id", i)
		}
		if seen[d.ID] {
			return nil, fmt.Errorf("enemy %q: duplicate id", d.ID)
		}
		seen[d.ID] = true
		if err := d.validate(); err != nil {
			return nil, fmt.Errorf("enemy %q: %w", d.ID, err)
		}
	}
	return defs, nil
}

func (d EnemyDef) validate() error {
	switch d.Behaviour {
	case BehaviourHopToward, BehaviourCharge, BehaviourPuff:
	default:
		return fmt.Errorf("unknown behaviour %q", d.Behaviour)
	}
	if err := d.validateStats("", d.HP, d.Damage, d.Speed); err != nil {
		return err
	}
	if d.TelegraphSeconds < MinTelegraph.Seconds() {
		return fmt.Errorf("telegraphSeconds %v is below the minimum %v", d.TelegraphSeconds, MinTelegraph.Seconds())
	}
	for i, dr := range d.Drops {
		if dr.Item == "" {
			return fmt.Errorf("drop %d: empty item", i)
		}
		if dr.Chance <= 0 || dr.Chance > 1 {
			return fmt.Errorf("drop %d %q: chance %v is outside (0, 1]", i, dr.Item, dr.Chance)
		}
	}
	if o := d.LittleExplorer; o != nil {
		hp, damage, speed := d.HP, d.Damage, d.Speed
		if o.HP != nil {
			hp = *o.HP
		}
		if o.Damage != nil {
			damage = *o.Damage
		}
		if o.Speed != nil {
			speed = *o.Speed
		}
		if err := d.validateStats("littleExplorer ", hp, damage, speed); err != nil {
			return err
		}
	}
	return nil
}

func (d EnemyDef) validateStats(prefix string, hp, damage int, speed float64) error {
	switch {
	case hp <= 0:
		return fmt.Errorf("%shp %d is not positive", prefix, hp)
	case damage <= 0:
		return fmt.Errorf("%sdamage %d is not positive", prefix, damage)
	case speed <= 0:
		return fmt.Errorf("%sspeed %v is not positive", prefix, speed)
	}
	return nil
}

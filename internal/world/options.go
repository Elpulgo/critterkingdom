package world

import "github.com/Elpulgo/critterkingdom/internal/combat"

// Option changes how NewScene sets up a scene.
type Option func(*Scene)

// WithEnemies gives the scene the enemy definitions that the maps' enemy
// objects name. Without it, a map with an enemy object cannot be loaded.
func WithEnemies(defs []combat.EnemyDef) Option {
	return func(s *Scene) {
		for _, d := range defs {
			s.enemyDefs[d.ID] = d
		}
	}
}

// WithEnemyDefeated gives the scene a sink that is called once for each
// enemy defeat.
func WithEnemyDefeated(sink func(combat.EnemyDefeated)) Option {
	return func(s *Scene) {
		s.onDefeat = sink
	}
}

// WithPreset spawns every enemy with the stats of the age preset p, as
// combat.EnemyDef.ForPreset gives them. Without it enemies have their
// Explorer, base, stats.
func WithPreset(p combat.Preset) Option {
	return func(s *Scene) {
		s.preset = p
	}
}

// WithPeaceful puts every enemy the scene spawns to sleep, so none of them
// ever moves or hurts the critter. They can still be defeated.
func WithPeaceful() Option {
	return func(s *Scene) {
		s.peaceful = true
	}
}

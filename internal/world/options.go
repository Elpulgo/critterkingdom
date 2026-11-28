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

// WithInteraction gives the scene a sink that is called once for each
// interaction the critter starts with a villager, on the tick it starts:
// by paw, or when a tapped walk reaches the villager. TakeInteraction still
// reports the same interactions.
func WithInteraction(sink func(Interaction)) Option {
	return func(s *Scene) {
		s.onInteract = sink
	}
}

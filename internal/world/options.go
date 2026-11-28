package world

import (
	"github.com/Elpulgo/critterkingdom/internal/anim"
	"github.com/Elpulgo/critterkingdom/internal/combat"
)

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

// Audio is what the scene needs from the game's sound: the music track of
// each map it enters, and once per tick whether the critter walked.
type Audio interface {
	// SetMusic asks for the map's music track, the map's music property;
	// "" when the map has none.
	SetMusic(track string)
	// Tick reports one tick of dt seconds, walking true when the critter
	// moved during it.
	Tick(dt float64, walking bool)
}

// WithAudio gives the scene a to play the music of each map it enters and to
// hear the critter's steps. Without it, or with a nil a, the scene is silent.
func WithAudio(a Audio) Option {
	return func(s *Scene) {
		s.audio = a
	}
}

// CharacterLoader loads the character with the given id, such as "fox": its
// part layout and the motions that animate it.
type CharacterLoader func(id string) (anim.Character, error)

// WithCharacters gives the scene the loader for the critter's character,
// which NewScene loads and animates. Without it, the critter has no parts and
// only its shadow is drawn.
func WithCharacters(load CharacterLoader) Option {
	return func(s *Scene) {
		s.loadCharacter = load
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

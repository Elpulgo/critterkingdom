package combat

import "github.com/Elpulgo/critterkingdom/internal/geom"

// KnockbackDistance is how many world pixels a hit pushes its target back.
const KnockbackDistance = 12.0

// Knockback returns a displacement of length KnockbackDistance pointing from
// attacker to target, or the zero vector when the two points are the same.
func Knockback(attacker, target geom.Vec) geom.Vec {
	return target.Sub(attacker).Normalized().Scale(KnockbackDistance)
}

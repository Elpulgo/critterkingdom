package world

import (
	"fmt"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

const (
	// enemyType is the Tiled object type of the place an enemy stands.
	enemyType = "enemy"
	// enemyProperty is the string property of an enemy object that holds the
	// enemy definition id, for example "sulky_slime".
	enemyProperty = "enemy"
)

// EnemyMarker is an enemy object of the objects layer: which enemy stands
// where when the map is entered.
type EnemyMarker struct {
	// ObjectID is the Tiled object id, for error messages.
	ObjectID int
	// EnemyID is the enemy definition id from the string property enemy.
	EnemyID string
	// Pos is the object's position in world pixels.
	Pos geom.Vec
}

// Enemies returns every enemy object of the objects layer, in layer order.
// The slice is a copy; changing it does not change the map.
func (m *Map) Enemies() []EnemyMarker {
	return append([]EnemyMarker(nil), m.enemies...)
}

// readEnemy turns the enemy object o into an EnemyMarker. It returns an error
// naming the object id when o has no non-empty string property enemy.
func readEnemy(o Object) (EnemyMarker, error) {
	id := o.Properties.String(enemyProperty)
	if id == "" {
		return EnemyMarker{}, fmt.Errorf("%s object %d: enemy has no string property %s", objectsLayer, o.ID, enemyProperty)
	}
	return EnemyMarker{ObjectID: o.ID, EnemyID: id, Pos: geom.Vec{X: o.X, Y: o.Y}}, nil
}

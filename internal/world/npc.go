package world

import (
	"fmt"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// npcType is the Tiled object type of a villager's spawn and wander area.
const npcType = "npc"

// NPCSpawn is where a villager starts on a map and the area it wanders in,
// read from an npc object of the objects layer.
type NPCSpawn struct {
	// ObjectID is the Tiled object id, unique on the map.
	ObjectID int
	// ID names the villager, from the string property id.
	ID string
	// Pos is where the villager's feet start, in world pixels: the centre of
	// the object's rectangle, or the point of a point object.
	Pos geom.Vec
	// Area is the rectangle the villager wanders in, in world pixels. A point
	// object has a zero-size Area at Pos: a villager that stands still.
	Area geom.Rect
}

// NPCs returns one spawn per npc object of the objects layer, in layer order.
// The slice is a copy; changing it does not change the map.
func (m *Map) NPCs() []NPCSpawn {
	return append([]NPCSpawn(nil), m.npcs...)
}

// loadNPC reads the npc object o into m. seen maps each npc id already read
// to its object id. It returns an error naming the object id for an npc
// without an id property or with an id already used on the map.
func (m *Map) loadNPC(o Object, seen map[string]int) error {
	id := o.Properties.String("id")
	if id == "" {
		return fmt.Errorf("%s object %d: npc has no string property id", objectsLayer, o.ID)
	}
	if first, dup := seen[id]; dup {
		return fmt.Errorf("%s object %d: npc id %q is already used by object %d", objectsLayer, o.ID, id, first)
	}
	seen[id] = o.ID
	area := o.Rect()
	m.npcs = append(m.npcs, NPCSpawn{
		ObjectID: o.ID,
		ID:       id,
		Pos:      area.Min.Add(area.Max).Scale(0.5),
		Area:     area,
	})
	return nil
}

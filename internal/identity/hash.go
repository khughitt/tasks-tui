package identity

import "hash/fnv"

// SlotCount is familiar-theme's SLOT_COUNT. FROZEN with the table below.
const SlotCount = 12

type HueSat struct {
	Hue int
	Sat int
}

// Slots is familiar's slot-hues.js table. FROZEN: a project pinned to slot 0 is the
// same warm orange here as on its pet. Changing a row changes every project's colour.
var Slots = [SlotCount]HueSat{
	{22, 62}, {40, 58}, {55, 42}, {95, 38}, {135, 45}, {172, 48},
	{212, 55}, {248, 45}, {272, 45}, {305, 45}, {350, 38}, {0, 6},
}

// FNV1a32 over UTF-8 bytes, familiar's fnv1a32. FROZEN: it decides every unpinned slot.
func FNV1a32(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

// AutoSlot is the unpinned slot for a project key.
func AutoSlot(key string) int { return int(FNV1a32(key) % SlotCount) }

package render

import (
	"github.com/LedFx/ledfx/pkg/color"
	"github.com/LedFx/ledfx/pkg/device"
	log "github.com/LedFx/ledfx/pkg/logger"
)

// A group of device's pixels. Effects render onto a pixel group.
type PixelGroup struct {
	Group      map[string]color.Pixels // the group of pixels. maps device id to pixels
	Order      []string                // defines the order of the pixels in the group
	Largest    string                  // the id of the largest pixel output in the group
	Smallest   string                  // the id of the smallest pixel output in the group
	LargestLen int                     // length of the largest pixel output in the group
	TotalLen   int                     // total number of pixels
}

// creates a pixel group for a slice of device IDs
func NewPixelGroup(devices map[string]*device.Device, order []string) (pg *PixelGroup, err error) {
	pg = new(PixelGroup)
	pg.Group = make(map[string]color.Pixels)
	if len(devices) == 0 {
		return
	}
	var largest, smallest string
	for id, d := range devices {
		// initialise size search variables
		if largest == "" {
			largest = id
		}
		if smallest == "" {
			smallest = id
		}
		// add pixels to group
		pg.Group[id] = make(color.Pixels, d.Config.PixelCount)
		if d.Config.PixelCount > pg.LargestLen {
			pg.LargestLen = d.Config.PixelCount
		}
		pg.TotalLen += d.Config.PixelCount
	}
	// determine largest and smallest
	for id, px := range pg.Group {
		if len(px) > len(pg.Group[largest]) {
			largest = id
		}
		if len(px) < len(pg.Group[smallest]) {
			smallest = id
		}
	}
	pg.Largest = largest
	pg.Smallest = smallest
	// Validate the ordering. We'll assume it's okay, and subject it to some test cases.
	// Order should be the set of device keys.
	// Order should have the same number as devices, not repeated, and all correspond to a device.
	allGood := true
Validation:
	for i, key := range order {
		// test if every id in order is a one of the devices
		if _, ok := devices[key]; !ok {
			allGood = false
			break Validation
		}
		// test if the key is unique
		for j, otherKey := range order {
			// dont compare key to itself
			if i == j {
				continue
			}
			if key == otherKey {
				allGood = false
				break Validation
			}
		}
	}

	// there should be the same number of keys in order as devices
	if len(devices) != len(order) {
		allGood = false
	}

	if allGood {
		pg.Order = order
		return
	}

	// just randomly add the devices to the order
	for id := range devices {
		pg.Order = append(pg.Order, id)
	}

	return
}

// Clones the pixels for a given id to all other pixel outputs.
func (pg *PixelGroup) CloneToAll(id string) {
	if _, ok := pg.Group[id]; !ok {
		log.Logger.WithField("context", "Pixel Group").Debugf("ID %s does not exist in this pixel group", id)
		return
	}
	cloneFrom := pg.Group[id]
	for cloneTo := range pg.Group {
		color.Interpolate(cloneFrom, pg.Group[cloneTo])
	}
}

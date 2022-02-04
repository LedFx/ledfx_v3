package audio

type Modulator struct {
	b Buffer

	vSurroundModifier int16
}

func (b Buffer) NewModulator() *Modulator {
	return &Modulator{
		b: b,
	}
}

func (m *Modulator) VirtualSurround() Buffer {
	modfn := func(mod int16) int16 {
		return mod + 1
	}
	for i := 0; i < 12; i++ {
		for i2 := 0; i2 < len(m.b); i2 += 2 {
			m.b[i2] -= m.vSurroundModifier
			if !(i2+1 >= len(m.b)) {
				m.b[i2+1] += m.vSurroundModifier
			}

			if m.vSurroundModifier >= 6 {
				modfn = func(mod int16) int16 {
					return mod - 1
				}
			} else if m.vSurroundModifier == 0 {
				modfn = func(mod int16) int16 {
					return mod + 1
				}
			}
			m.vSurroundModifier = modfn(m.vSurroundModifier)
		}
	}

	return m.b
}

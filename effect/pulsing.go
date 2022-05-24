package effect

import (
	"fmt"
	"ledfx/color"
)

type PulsingEffect struct{}

func (e *PulsingEffect) AssembleFrame(p *color.Pixels)             {}
func (e *PulsingEffect) Initialize(id string, config EffectConfig) {}
func (e *PulsingEffect) ConfigUpdated()                            {}
func (e *PulsingEffect) AudioUpdated()                             {}

type Speaker interface {
	Speak()
}

type config struct {
	id string
}

type studentConfig struct {
	config
	studentNumber int
}

type Person struct {
	Name   string
	Age    int
	Config config
}

type Student struct {
	Person
	Subject string
	Config  studentConfig
}

func (p *Person) Speak() {
	fmt.Println("My name is" + p.Name)
}

func (s *Student) Speak() {
	fmt.Println("My name is" + s.Name + "and i study" + s.Subject)
}

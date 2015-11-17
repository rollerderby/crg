package leagues

import (
	"errors"

	"github.com/rollerderby/crg/statemanager"
)

type Person struct {
	id              string
	name            string
	legalName       string
	insuranceNumber string
	number          string
	stateIDs        map[string]string
}

var persons = make(map[string]*Person)
var errPersonNotFound = errors.New("Person Not Found")

func blankPerson(id string) *Person {
	p := &Person{
		stateIDs: make(map[string]string),
	}

	base := "Leagues.Person(" + id + ")"
	p.stateIDs["id"] = base + ".ID"
	p.stateIDs["name"] = base + ".Name"
	p.stateIDs["legalName"] = base + ".LegalName"
	p.stateIDs["insuranceNumber"] = base + ".InsuranceNumber"
	p.stateIDs["number"] = base + ".Number"

	p.SetID(id)
	p.SetName("")
	p.SetLegalName("")
	p.SetInsuranceNumber("")
	p.SetNumber("")

	persons[id] = p

	return p
}

func NewPerson(id, name, legalName, insuranceNumber, number string) *Person {
	p := blankPerson(id)
	p.SetName(name)
	p.SetLegalName(legalName)
	p.SetInsuranceNumber(insuranceNumber)
	p.SetNumber(number)

	return p
}

func (p *Person) ID() string { return p.id }
func (p *Person) SetID(v string) error {
	p.id = v
	return statemanager.StateUpdate(p.stateIDs["id"], v)
}

func (p *Person) Name() string { return p.name }
func (p *Person) SetName(v string) error {
	p.name = v
	return statemanager.StateUpdate(p.stateIDs["name"], v)
}

func (p *Person) LegalName() string { return p.legalName }
func (p *Person) SetLegalName(v string) error {
	p.legalName = v
	return statemanager.StateUpdate(p.stateIDs["legalName"], v)
}

func (p *Person) InsuranceNumber() string { return p.insuranceNumber }
func (p *Person) SetInsuranceNumber(v string) error {
	p.insuranceNumber = v
	return statemanager.StateUpdate(p.stateIDs["insuranceNumber"], v)
}

func (p *Person) Number() string { return p.number }
func (p *Person) SetNumber(v string) error {
	p.number = v
	return statemanager.StateUpdate(p.stateIDs["number"], v)
}

/* Helper functions to find the Person for RegisterUpdaters */
func findPerson(k string) *Person {
	ids := statemanager.ParseIDs(k)
	id := ids[0]

	p, ok := persons[id]
	if !ok {
		p = blankPerson(id)
		persons[id] = p
	}
	return p
}

func personSetID(k, v string) error {
	if p := findPerson(k); p != nil {
		p.SetID(v)
		return nil
	}
	return errPersonNotFound
}
func personSetName(k, v string) error {
	if p := findPerson(k); p != nil {
		p.SetName(v)
		return nil
	}
	return errPersonNotFound
}
func personSetLegalName(k, v string) error {
	if p := findPerson(k); p != nil {
		p.SetLegalName(v)
		return nil
	}
	return errPersonNotFound
}
func personSetInsuranceNumber(k, v string) error {
	if p := findPerson(k); p != nil {
		p.SetInsuranceNumber(v)
		return nil
	}
	return errPersonNotFound
}
func personSetNumber(k, v string) error {
	if p := findPerson(k); p != nil {
		p.SetNumber(v)
		return nil
	}
	return errPersonNotFound
}

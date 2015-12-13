// Copyright 2015 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

package leagues

import (
	"errors"

	"github.com/rollerderby/crg/state"
	"github.com/rollerderby/crg/utils"
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

// NewPerson creates a new person (unattached to any leagues at this point)
func NewPerson(id, name, legalName, insuranceNumber, number string) *Person {
	p := blankPerson(id)
	p.SetName(name)
	p.SetLegalName(legalName)
	p.SetInsuranceNumber(insuranceNumber)
	p.SetNumber(number)

	return p
}

// ID returns the id
func (p *Person) ID() string { return p.id }

// SetID sets the ID to `v`
func (p *Person) SetID(v string) error {
	p.id = v
	state.SetStateString(p.stateIDs["id"], v)
	return nil
}

func (p *Person) Name() string { return p.name }
func (p *Person) SetName(v string) error {
	p.name = v
	state.SetStateString(p.stateIDs["name"], v)
	return nil
}

func (p *Person) LegalName() string { return p.legalName }
func (p *Person) SetLegalName(v string) error {
	p.legalName = v
	state.SetStateString(p.stateIDs["legalName"], v)
	return nil
}

func (p *Person) InsuranceNumber() string { return p.insuranceNumber }
func (p *Person) SetInsuranceNumber(v string) error {
	p.insuranceNumber = v
	state.SetStateString(p.stateIDs["insuranceNumber"], v)
	return nil
}

func (p *Person) Number() string { return p.number }
func (p *Person) SetNumber(v string) error {
	p.number = v
	state.SetStateString(p.stateIDs["number"], v)
	return nil
}

/* Helper functions to find the Person for RegisterUpdaters */
func findPerson(k string) *Person {
	ids := utils.ParseIDs(k)
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

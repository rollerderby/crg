package leagues

import "github.com/rollerderby/crg/statemanager"

type league struct {
	name string
	// members []*leagueMember
	// teams   []*team
}

// Initialize the leagues subsystem
func Initialize() {
	statemanager.RegisterPatternUpdaterString("Leagues.Person(*).ID", 0, personSetID)
	statemanager.RegisterPatternUpdaterString("Leagues.Person(*).Name", 0, personSetName)
	statemanager.RegisterPatternUpdaterString("Leagues.Person(*).LegalName", 0, personSetLegalName)
	statemanager.RegisterPatternUpdaterString("Leagues.Person(*).InsuranceNumber", 0, personSetInsuranceNumber)
	statemanager.RegisterPatternUpdaterString("Leagues.Person(*).Number", 0, personSetNumber)
}

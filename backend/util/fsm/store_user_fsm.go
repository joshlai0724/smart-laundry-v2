package fsmutil

import "github.com/looplab/fsm"

const (
	StoreUserStateActive   string = "active"
	StoreUserStateArchived string = "archived"
	InitStoreUserState     string = StoreUserStateActive

	StoreUserEventDeactive  string = "deactive"
	StoreUserEventEnable    string = "enable"
	StoreUserEventCashTopUp string = "cash_top_up"
)

func NewStoreUserFSM(initState string) *fsm.FSM {
	return fsm.NewFSM(
		initState,
		fsm.Events{
			{Name: StoreUserEventDeactive, Src: []string{StoreUserStateActive}, Dst: StoreUserStateArchived},
			{Name: StoreUserEventEnable, Src: []string{StoreUserStateArchived}, Dst: StoreUserStateActive},
			{Name: StoreUserEventCashTopUp, Src: []string{StoreUserStateActive}, Dst: StoreUserStateActive},
		},
		map[string]fsm.Callback{},
	)
}

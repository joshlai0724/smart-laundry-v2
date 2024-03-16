package fsmutil

import "github.com/looplab/fsm"

const (
	StoreStateActive   string = "active"
	StoreStateArchived string = "archived"
	InitStoreState     string = StoreStateActive

	StoreEventDeactive string = "deactive"
	StoreEventEnable   string = "enable"
)

func NewStoreFSM(initState string) *fsm.FSM {
	return fsm.NewFSM(
		initState,
		fsm.Events{
			{Name: StoreEventDeactive, Src: []string{StoreStateActive}, Dst: StoreStateArchived},
			{Name: StoreEventEnable, Src: []string{StoreStateArchived}, Dst: StoreStateActive},
		},
		map[string]fsm.Callback{},
	)
}

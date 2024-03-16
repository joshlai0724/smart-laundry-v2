package fsmutil

import "github.com/looplab/fsm"

const (
	UserStateActive string = "active"
	UserStateLocked string = "locked"
	InitUserState   string = UserStateActive

	UserEventLogin                     string = "login"
	UserEventResetPassword             string = "reset_password"
	UserEventChangePassword            string = "change_password"
	UserEventPasswordErrorTooManyTimes string = "password_error_too_many_times"
)

func NewUserFSM(initState string) *fsm.FSM {
	return fsm.NewFSM(
		initState,
		fsm.Events{
			{Name: UserEventLogin, Src: []string{UserStateActive}, Dst: UserStateActive},
			{Name: UserEventResetPassword, Src: []string{UserStateActive, UserStateLocked}, Dst: UserStateActive},
			{Name: UserEventChangePassword, Src: []string{UserStateActive, UserStateLocked}, Dst: UserStateActive},
			{Name: UserEventPasswordErrorTooManyTimes, Src: []string{UserStateActive}, Dst: UserStateLocked},
		},
		map[string]fsm.Callback{},
	)
}

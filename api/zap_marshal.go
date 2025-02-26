package api

import (
	"fmt"

	"go.uber.org/zap/zapcore"
)

type (
	Users         []User
	SlotifyGroups []SlotifyGroup
)

// MarshalLogArray implements the zapcore.ArrayMarshaler interface.
func (u Users) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for _, user := range u {
		if err := arr.AppendObject(user); err != nil {
			return err
		}
	}
	return nil
}

// MarshalLogObject implements zapcore.ObjectMarshaler.

func (u UserCreate) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("email", string(u.Email))
	enc.AddString("firstName", u.FirstName)
	enc.AddString("lastName", u.LastName)
	return nil
}

func (u User) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	userCreate := UserCreate{
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
	}
	if err := userCreate.MarshalLogObject(enc); err != nil {
		return fmt.Errorf("failed to marshal User obj: %v", err.Error())
	}
	enc.AddUint32("id", u.Id)
	return nil
}

// MarshalLogObject implements zapcore.ObjectMarshaler.
func (sgc SlotifyGroupCreate) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("name", sgc.Name)
	return nil
}

// MarshalLogArray implements the zapcore.ArrayMarshaler interface.
func (sgs SlotifyGroups) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for _, slotifyGroup := range sgs {
		if err := arr.AppendObject(slotifyGroup); err != nil {
			return err
		}
	}
	return nil
}

func (sg SlotifyGroup) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	slotifyGroupCreate := SlotifyGroupCreate{
		Name: sg.Name,
	}
	if err := slotifyGroupCreate.MarshalLogObject(enc); err != nil {
		return fmt.Errorf("failed to marshal SlotifyGroup obj: %v", err.Error())
	}
	enc.AddUint32("id", sg.Id)
	return nil
}

func (pai PostAPIInvitesJSONRequestBody) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("message", pai.Message)
	enc.AddUint32("toUserID", pai.ToUserID)
	enc.AddUint32("slotifyGroupID", pai.SlotifyGroupID)
	enc.AddTime("expiryDate", pai.ExpiryDate.Time)
	enc.AddTime("createdAt", pai.CreatedAt)
	return nil
}

func (paiBody PatchAPIInvitesInviteIDJSONRequestBody) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("message", paiBody.Message)
	return nil
}

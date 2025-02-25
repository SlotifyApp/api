package api

import (
	"fmt"

	"go.uber.org/zap/zapcore"
)

type (
	Users []User
	Teams []Team
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
func (t TeamCreate) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("name", t.Name)
	return nil
}

// MarshalLogArray implements the zapcore.ArrayMarshaler interface.
func (t Teams) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for _, team := range t {
		if err := arr.AppendObject(team); err != nil {
			return err
		}
	}
	return nil
}

func (t Team) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	teamCreate := TeamCreate{
		Name: t.Name,
	}
	if err := teamCreate.MarshalLogObject(enc); err != nil {
		return fmt.Errorf("failed to marshal Team obj: %v", err.Error())
	}
	enc.AddUint32("id", t.Id)
	return nil
}

func (tp GetAPITeamsParams) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	name := ""
	if tp.Name != nil {
		name = *tp.Name
	}
	enc.AddString("name", name)
	return nil
}

func (g GetAPIGroupsParams) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	name := ""
	if g.Name != nil {
		name = *g.Name
	} else {
		name = "nil"
	}
	enc.AddString("name", name)
	return nil
}

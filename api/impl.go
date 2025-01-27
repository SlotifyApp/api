package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	PrepareStmtFail = "failed to prepare sql stmt"
	QueryDBFail     = "failed to query database"
)

// ensure that we've conformed to the `ServerInterface` with a compile-time check.
var _ ServerInterface = (*Server)(nil)

// sendError wraps sending of an error in the Error format, and
// handling the failure to marshal that.
func sendError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(message)
}

// Set JSON content-type header and send response.
func SetHeaderAndWriteResponse(w http.ResponseWriter, code int, encode any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(encode); err != nil {
		sendError(w, http.StatusInternalServerError, "failed to encode JSON")
	}
}

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
	enc.AddInt("id", u.Id)
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
	enc.AddInt("id", t.Id)
	return nil
}

func (tp GetTeamsParams) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	name := ""
	if tp.Name != nil {
		name = *tp.Name
	}
	enc.AddString("name", name)
	return nil
}

// (GET /healthcheck).
func (s Server) GetHealthcheck(w http.ResponseWriter, _ *http.Request) {
	resp := "Healthcheck Successful!"
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.Logger.Error("Failed to encode JSON", zap.String("body", resp))
		sendError(w, http.StatusInternalServerError, "Failed to encode JSON")
		return
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/labstack/echo/v4"
)

func newSerializer() echo.JSONSerializer {
	return sonicJSONSerializer{
		api: sonic.ConfigFastest,
	}
}

// idea of using Echo's serialized was from from https://github.com/tomruk/fj4echo (MIT Licence)
// sonicJSONSerializer implements JSON encoding using github.com/bytedance/sonic
type sonicJSONSerializer struct {
	api sonic.API
}

// Serialize converts an interface into a json and writes it to the response.
// You can optionally use the indent parameter to produce pretty JSONs.
func (s sonicJSONSerializer) Serialize(c echo.Context, i interface{}, indent string) error {
	enc := s.api.NewEncoder(c.Response())
	return enc.Encode(i)
}

// Deserialize reads a JSON from a request body and converts it into an interface.
func (s sonicJSONSerializer) Deserialize(c echo.Context, i interface{}) error {
	err := s.api.NewDecoder(c.Request().Body).Decode(i)
	if ute, ok := err.(*json.UnmarshalTypeError); ok {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Unmarshal type error: expected=%v, got=%v, field=%v, offset=%v", ute.Type, ute.Value, ute.Field, ute.Offset)).SetInternal(err)
	} else if se, ok := err.(*json.SyntaxError); ok {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Syntax error: offset=%v, error=%v", se.Offset, se.Error())).SetInternal(err)
	}
	return err
}

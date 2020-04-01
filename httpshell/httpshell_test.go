package httpshell_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/flex"
	"github.com/jackc/flex/httpshell"
	"github.com/stretchr/testify/assert"
)

func TestJSONHandlerEmptyResponse(t *testing.T) {
	shell := flex.NewShell()
	shell.Register(&flex.Command{
		Name: "nop",
		ParamsType: flex.NewType(func(tb flex.TypeBuilder) {
		}),
		ExecJSONFunc: func(ctx context.Context, params *flex.Record) ([]byte, error) {
			return nil, nil
		},
	})

	jsonHandler := &httpshell.JSONHandler{
		Shell:       shell,
		CommandName: "nop",
		BuildParams: func(*http.Request) (map[string]interface{}, error) { return nil, nil },
		HandleError: func(w http.ResponseWriter, r *http.Request, err error) { return },
	}

	req := httptest.NewRequest("GET", "http://example.com/", nil)
	w := httptest.NewRecorder()
	jsonHandler.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	assert.Len(t, body, 0)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

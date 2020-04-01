package httpshell_test

import (
	"context"
	"fmt"
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
		Shell:         shell,
		CommandName:   "nop",
		ParamsBuilder: func(*http.Request) (map[string]interface{}, error) { return nil, nil },
		ErrorHandler:  func(w http.ResponseWriter, r *http.Request, err error) { return },
	}

	req := httptest.NewRequest("GET", "http://example.com/", nil)
	w := httptest.NewRecorder()
	jsonHandler.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	assert.Len(t, body, 0)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestJSONHandlerWithoutBuildParams(t *testing.T) {
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
		Shell:        shell,
		CommandName:  "nop",
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) { return },
	}

	req := httptest.NewRequest("GET", "http://example.com/", nil)
	w := httptest.NewRecorder()
	jsonHandler.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	assert.Len(t, body, 0)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestJSONHandlerErrorWithoutErrorHandlerPanics(t *testing.T) {
	shell := flex.NewShell()
	shell.Register(&flex.Command{
		Name: "nop",
		ParamsType: flex.NewType(func(tb flex.TypeBuilder) {
		}),
		ExecJSONFunc: func(ctx context.Context, params *flex.Record) ([]byte, error) {
			return nil, fmt.Errorf("something went wrong")
		},
	})

	jsonHandler := &httpshell.JSONHandler{
		Shell:       shell,
		CommandName: "nop",
	}

	req := httptest.NewRequest("GET", "http://example.com/", nil)
	w := httptest.NewRecorder()

	assert.PanicsWithError(t, "missing error handler: nop: exec: something went wrong", func() {
		jsonHandler.ServeHTTP(w, req)
	})
}

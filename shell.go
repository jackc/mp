package flex

import (
	"context"
	"encoding/json"
	"fmt"
)

type Shell struct {
	commands map[string]*Command
}

func NewShell() *Shell {
	return &Shell{
		commands: make(map[string]*Command),
	}
}

func (s *Shell) Register(f *Command) {
	s.commands[f.Name] = f
}

func (s *Shell) Commands() map[string]*Command {
	return s.commands
}

func (s *Shell) Exec(ctx context.Context, name string, params map[string]interface{}) (map[string]interface{}, error) {
	cmd, ok := s.commands[name]
	if !ok {
		return nil, fmt.Errorf("command not found: %s", name)
	}

	return cmd.Exec(ctx, params)
}

func (s *Shell) ExecJSON(ctx context.Context, name string, params map[string]interface{}) ([]byte, error) {
	cmd, ok := s.commands[name]
	if !ok {
		return nil, fmt.Errorf("command not found: %s", name)
	}

	return cmd.ExecJSON(ctx, params)
}

type Command struct {
	Name         string
	ParamsType   *Type
	ExecFunc     func(ctx context.Context, params *Record) (map[string]interface{}, error)
	ExecJSONFunc func(ctx context.Context, params *Record) ([]byte, error)
}

func (cmd *Command) Exec(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	parsedParams := cmd.ParamsType.New(params)
	if parsedParams.Errors() != nil {
		return nil, fmt.Errorf("%s: failed to parse params: %w", cmd.Name, parsedParams.Errors())
	}

	if cmd.ExecFunc != nil {
		response, err := cmd.ExecFunc(ctx, parsedParams)
		if err != nil {
			return nil, fmt.Errorf("%s: ExecFunc: %w", cmd.Name, err)
		}
		return response, nil
	}

	if cmd.ExecJSONFunc != nil {
		buf, err := cmd.ExecJSONFunc(ctx, parsedParams)
		if err != nil {
			return nil, fmt.Errorf("%s: ExecJSONFunc: %w", cmd.Name, err)
		}

		if buf == nil {
			return nil, nil
		}

		var response map[string]interface{}
		err = json.Unmarshal(buf, &response)
		if err != nil {
			return nil, fmt.Errorf("%s: json.Unmarshal: %w", cmd.Name, err)
		}
		return response, nil
	}

	return nil, fmt.Errorf("%s: missing function", cmd.Name)
}

func (cmd *Command) ExecJSON(ctx context.Context, params map[string]interface{}) ([]byte, error) {
	parsedParams := cmd.ParamsType.New(params)
	if parsedParams.Errors() != nil {
		return nil, fmt.Errorf("%s: failed to parse params: %w", cmd.Name, parsedParams.Errors())
	}

	if cmd.ExecJSONFunc != nil {
		buf, err := cmd.ExecJSONFunc(ctx, parsedParams)
		if err != nil {
			return nil, fmt.Errorf("%s: ExecJSONFunc: %w", cmd.Name, err)
		}
		return buf, nil
	}

	if cmd.ExecFunc != nil {
		response, err := cmd.ExecFunc(ctx, parsedParams)
		if err != nil {
			return nil, fmt.Errorf("%s: ExecFunc: %w", cmd.Name, err)
		}

		if response == nil {
			return nil, nil
		}

		buf, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("%s: json.Marshal: %w", cmd.Name, err)
		}
		return buf, nil
	}

	return nil, fmt.Errorf("%s: missing function", cmd.Name)
}

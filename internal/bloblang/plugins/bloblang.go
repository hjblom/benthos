package plugins

import (
	"github.com/Jeffail/benthos/v3/internal/bloblang/parser"
	"github.com/Jeffail/benthos/v3/internal/bloblang/query"
	"github.com/Jeffail/benthos/v3/lib/message"
)

// Register adds any native Bloblang methods and functions to the global sets
// that aren't defined within the query package.
func Register() error {
	dynamicBloblangParserContext := parser.Context{
		Functions: query.AllFunctions.OnlyPure().NoMessage(),
		Methods:   query.AllMethods,
	}

	return query.AllMethods.Add(
		query.NewMethodSpec(
			"bloblang", "Executes an argument Bloblang mapping on the target. This method can be used in order to execute dynamic mappings. Functions that interact with the environment, such as `file` and `env`, or that access message information directly, such as `content` or `json`, are not enabled for dynamic Bloblang mappings.",
		).InCategory(
			query.MethodCategoryParsing, "",
			query.NewExampleSpec(
				"",
				"root.body = this.body.bloblang(this.mapping)",
				`{"body":{"foo":"hello world"},"mapping":"root.foo = this.foo.uppercase()"}`,
				`{"body":{"foo":"HELLO WORLD"}}`,
				`{"body":{"foo":"hello world 2"},"mapping":"root.foo = this.foo.capitalize()"}`,
				`{"body":{"foo":"Hello World 2"}}`,
			),
		).Beta().Param(query.ParamString("mapping", "The mapping to execute.")),
		func(target query.Function, args *query.ParsedParams) (query.Function, error) {
			mappingStr, err := args.FieldString("mapping")
			if err != nil {
				return nil, err
			}
			exec, parserErr := parser.ParseMapping(dynamicBloblangParserContext, "", mappingStr)
			if parserErr != nil {
				return nil, parserErr
			}
			return query.ClosureFunction("method bloblang", func(ctx query.FunctionContext) (interface{}, error) {
				v, err := target.Exec(ctx)
				if err != nil {
					return nil, err
				}
				return exec.Exec(query.FunctionContext{
					Vars:     map[string]interface{}{},
					Maps:     exec.Maps(),
					MsgBatch: message.New(nil),
				}.WithValue(v))
			}, target.QueryTargets), nil
		},
	)
}

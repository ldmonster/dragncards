package functions

import (
	"errors"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/game"
)

// PluginCardFunction retrieves a plugin card by plugin and card id from context.
type PluginCardFunction struct{}

func (*PluginCardFunction) Name() string { return PluginCardFunctionName }
func (*PluginCardFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(PluginCardFunctionName, args, 2); err != nil {
		return nil, err
	}
	pluginID, ok := args[0].(string)
	if !ok {
		return nil, errors.New("plugin_card first arg must be string")
	}
	cardID, ok := args[1].(string)
	if !ok {
		return nil, errors.New("plugin_card second arg must be string")
	}
	if ctx == nil || ctx.Vars == nil {
		return nil, errors.New("plugin_card requires evaluation context with Var plugin_cards")
	}
	all, ok := ctx.Vars["plugin_cards"].(map[string]map[string]any)
	if !ok {
		return nil, errors.New("plugin_card requires variable plugin_cards to be map[string]map[string]any")
	}
	pluginTable, exists := all[pluginID]
	if !exists {
		return nil, nil
	}
	val, exists := pluginTable[cardID]
	if !exists {
		return nil, nil
	}
	return val, nil
}

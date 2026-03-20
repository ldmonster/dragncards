package game

// GameUI is the aggregate root for a running game session.
// It holds all runtime state: players, card entities, layout groups/stacks,
// and the append-only action log used for replay.
type GameUI struct {
	Slug       string            `json:"slug"`
	Players    []string          `json:"players"`
	Actions    [][]byte          `json:"actions"`
	Seats      map[string]string `json:"seats"`
	Spectators map[string]bool   `json:"spectators"`

	// Runtime game state — populated by the DSL evaluator.
	Cards  map[string]*Card       `json:"cards"`
	Stacks map[string]*Stack      `json:"stacks"`
	Groups map[string]*Group      `json:"groups"`
	Infos  map[string]*PlayerInfo `json:"player_infos"`
}

func NewGameUI(slug string) *GameUI {
	return &GameUI{
		Slug:       slug,
		Players:    []string{},
		Actions:    [][]byte{},
		Seats:      map[string]string{},
		Spectators: map[string]bool{},
		Cards:      map[string]*Card{},
		Stacks:     map[string]*Stack{},
		Groups:     map[string]*Group{},
		Infos:      map[string]*PlayerInfo{},
	}
}

// EvalContext carries DSL execution context, including state variables and previous result.
type EvalContext struct {
	Game *GameUI
	Vars map[string]any
	Prev any
	Eval func(card *Card, expr any) (any, error)
}

func NewEvalContext(gameUI *GameUI) *EvalContext {
	return &EvalContext{Game: gameUI, Vars: map[string]any{}, Prev: nil, Eval: nil}
}

func (c *EvalContext) Card(id string) *Card {
	if c == nil || c.Game == nil || c.Game.Cards == nil {
		return nil
	}
	return c.Game.Cards[id]
}

func (c *EvalContext) Stack(id string) *Stack {
	if c == nil || c.Game == nil || c.Game.Stacks == nil {
		return nil
	}
	return c.Game.Stacks[id]
}

func (c *EvalContext) Group(id string) *Group {
	if c == nil || c.Game == nil || c.Game.Groups == nil {
		return nil
	}
	return c.Game.Groups[id]
}

func (c *EvalContext) PlayerInfo(id string) *PlayerInfo {
	if c == nil || c.Game == nil || c.Game.Infos == nil {
		return nil
	}
	return c.Game.Infos[id]
}

func (g *GameUI) AddPlayer(playerID string) {
	if playerID == "" {
		return
	}
	for _, p := range g.Players {
		if p == playerID {
			return
		}
	}
	g.Players = append(g.Players, playerID)
}

func (g *GameUI) AddAction(payload []byte) {
	if len(payload) == 0 {
		return
	}
	g.Actions = append(g.Actions, payload)
}

// ResetActions clears the action log (used by reset_game).
func (g *GameUI) ResetActions() {
	g.Actions = [][]byte{}
}

func (g *GameUI) ResetSeats() {
	g.Seats = map[string]string{}
}

func (g *GameUI) ResetSpectators() {
	g.Spectators = map[string]bool{}
}

func (g *GameUI) SetSeat(playerID, seat string) {
	if playerID == "" {
		return
	}
	if g.Seats == nil {
		g.Seats = map[string]string{}
	}
	if seat == "" {
		delete(g.Seats, playerID)
		return
	}
	g.Seats[playerID] = seat
}

func (g *GameUI) SetSpectator(playerID string, spectator bool) {
	if playerID == "" {
		return
	}
	if g.Spectators == nil {
		g.Spectators = map[string]bool{}
	}
	if spectator {
		g.Spectators[playerID] = true
	} else {
		delete(g.Spectators, playerID)
	}
}

// PutCard upserts a Card into the game state.
func (g *GameUI) PutCard(c *Card) {
	if c == nil || c.ID == "" {
		return
	}
	if g.Cards == nil {
		g.Cards = map[string]*Card{}
	}
	g.Cards[c.ID] = c
}

// PutStack upserts a Stack into the game state.
func (g *GameUI) PutStack(s *Stack) {
	if s == nil || s.ID == "" {
		return
	}
	if g.Stacks == nil {
		g.Stacks = map[string]*Stack{}
	}
	g.Stacks[s.ID] = s
}

// PutGroup upserts a Group into the game state.
func (g *GameUI) PutGroup(gr *Group) {
	if gr == nil || gr.ID == "" {
		return
	}
	if g.Groups == nil {
		g.Groups = map[string]*Group{}
	}
	g.Groups[gr.ID] = gr
}

// PutPlayerInfo upserts a PlayerInfo entry.
func (g *GameUI) PutPlayerInfo(p *PlayerInfo) {
	if p == nil || p.ID == "" {
		return
	}
	if g.Infos == nil {
		g.Infos = map[string]*PlayerInfo{}
	}
	g.Infos[p.ID] = p
}

func (g *GameUI) RemoveCard(cardID string) {
	card := g.Cards[cardID]
	if card == nil {
		return
	}
	if stack := g.Stacks[card.StackID]; stack != nil {
		stack.RemoveCard(cardID)
	}
	delete(g.Cards, cardID)
}

func (g *GameUI) GetCard(cardID string) *Card {
	if g.Cards == nil {
		return nil
	}
	return g.Cards[cardID]
}

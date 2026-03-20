package game

// GameUI is the aggregate root for a running game session.
// It holds all runtime state: players, card entities, layout groups/stacks,
// and the append-only action log used for replay.
type GameUI struct {
	Slug    string   `json:"slug"`
	Players []string `json:"players"`
	Actions [][]byte `json:"actions"`

	// Runtime game state — populated by the DSL evaluator.
	Cards   map[string]*Card       `json:"cards"`
	Stacks  map[string]*Stack      `json:"stacks"`
	Groups  map[string]*Group      `json:"groups"`
	Infos   map[string]*PlayerInfo `json:"player_infos"`
}

func NewGameUI(slug string) *GameUI {
	return &GameUI{
		Slug:    slug,
		Players: []string{},
		Actions: [][]byte{},
		Cards:   map[string]*Card{},
		Stacks:  map[string]*Stack{},
		Groups:  map[string]*Group{},
		Infos:   map[string]*PlayerInfo{},
	}
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

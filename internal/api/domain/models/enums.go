package models

type GameName string

const (
	GameMagicTheGathering GameName = "magic the gathering"
	GamePokemon           GameName = "pokemon"
	GameFleshAndBlood     GameName = "flesh and blood"
	GameYugioh            GameName = "yugioh"
	GameDigimon           GameName = "digimon"
	GameOnePiece          GameName = "one piece"
	GameLorcana           GameName = "lorcana"
)

type TournamentScope string

const (
	ScopeLocal     TournamentScope = "LOCAL"
	ScopeQualifier TournamentScope = "QUALIFIER"
	ScopeMajor     TournamentScope = "MAJOR"
	ScopePremier   TournamentScope = "PREMIER"
)

type ResultName string

const (
	ResultWin  ResultName = "win"
	ResultDraw ResultName = "draw"
	ResultLose ResultName = "lose"
)

type MatchPhase string

const (
	PhaseSwiss  MatchPhase = "SWISS"
	PhaseTop8   MatchPhase = "TOP_8"
	PhaseTop4   MatchPhase = "TOP_4"
	PhaseFinals MatchPhase = "FINALS"
)

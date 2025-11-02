package shared

// Elemento: representa um item no mapa.
type Elemento struct {
	Simbolo  rune
	Cor      int
	CorFundo int
	Tangivel bool
}

// Jogador: representa o estado de um jogador.
type Jogador struct {
	ID                   int
	Elemento             Elemento
	PosX, PosY           int
	UltimoVisitado       Elemento
	Vidas                int
	UltimoSequenceNumber int64
}

// EstadoJogo: o que o servidor vai enviar para o cliente.
type EstadoJogo struct {
	Mapa      [][]Elemento
	Jogadores map[int]Jogador
}

// MoverRPC: contém os argumentos para a chamada RPC de movimento.
type MoverRPC struct {
	JogadorID      int
	Tecla          rune
	SequenceNumber int64
}

// RespostaServidorRPC: é a resposta do servidor a um novo jogador.
type RespostaServidorRPC struct {
	JogadorID int
	Estado    EstadoJogo
}

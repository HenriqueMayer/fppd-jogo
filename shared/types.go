// File shared/types.go para conter as definições de **structs** que tanto o cliente quanto
// o servidor irão utilizar. É tipo um contrato, ambos sabem o que entra e o que sai.

package shared

// Lembrando que para o RPC funcionar, todas as structs precisam ter os campos exportados
// - iniciar com letra maiúscula

// Elemento: representa um item no mapa.
type Elemento struct {
	Simbolo  rune // from wikipedia: "the rune type is an alias for the int32 type and is used to represent a single Unicode code point"
	Cor      int
	CorFundo int
	Tangivel bool
}

// Jogador: representa o estado de um jogador.
type Jogador struct {
	ID             int
	Elemento       Elemento // Representa o jogador no mapa
	PosX, PosY     int
	UltimoVisitado Elemento // O elemento que estava na posição antes do jogador se mover para lá
	Vidas          int
}

// EstadoDoJogo: o que o servidor vai enviar para o cliente.
type EstadoJogo struct {
	Mapa      [][]Elemento    // Uma matriz para representar o mapa do Jogo. "Onde as paredes ficam? Vegetação? Espaços vazios?"
	Jogadores map[int]Jogador // Um mapa de jogadores, esse muda conforme os jogadores. "Quem está jogando? Onde estão?"
}

// MoverRPC: contém os argumentos para a chamada RPC de movimento.
// É a anotação do movimento que o jogador quer fazer.
type MoverRPC struct {
	JogadorID      int
	Tecla          rune
	SequenceNumber int64 // Número para identificar a ordem das requisições
}

// RespostaServidorRPC: é a resposta do servidor a um novo jogador.
// É usada uma única vez. Você entra e recebe um ID e o estado atual do jogo.
type RespostaServidorRPC struct {
	JogadorID int
	Estado    EstadoJogo
}

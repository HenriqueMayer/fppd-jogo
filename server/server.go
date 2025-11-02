package main

import (
	"bufio"
	"log"
	"net"
	"net/rpc"
	"os"
	"sync"

	"jogo/shared"
)

type JogoServidor struct {
	mutex     sync.Mutex
	estado    shared.EstadoJogo
	proximoID int
}

// carrega o mapa de um arquivo txt
func (js *JogoServidor) jogoCarregarMapaLadoServidor(caminhoArquivo string) error {
	arq, err := os.Open(caminhoArquivo)
	if err != nil {
		return err
	}
	defer arq.Close()
	scanner := bufio.NewScanner(arq)

	for scanner.Scan() {
		linha := scanner.Text()
		var linhaElems []shared.Elemento
		for _, ch := range linha {
			var e shared.Elemento
			switch ch {
			case '▤':
				e = shared.Elemento{Simbolo: ch, Tangivel: true}
			case '♣':
				e = shared.Elemento{Simbolo: ch, Tangivel: false}
			default:
				e = shared.Elemento{Simbolo: ' ', Tangivel: false}
			}
			linhaElems = append(linhaElems, e)
		}
		js.estado.Mapa = append(js.estado.Mapa, linhaElems)
	}
	return scanner.Err()
}

func NovoJogoServidor() *JogoServidor {
	js := &JogoServidor{proximoID: 1}
	js.estado.Jogadores = make(map[int]shared.Jogador)
	if err := js.jogoCarregarMapaLadoServidor("game/mapa.txt"); err != nil {
		log.Fatalf("Nao consegui carregar o mapa: %v", err)
	}
	log.Println("Mapa carregado")
	return js
}

// -- Metodos RPC --

func (js *JogoServidor) RegistrarNovoJogador(args *struct{}, reply *shared.RespostaServidorRPC) error {
	js.mutex.Lock()
	defer js.mutex.Unlock()

	novoID := js.proximoID
	novoJogador := shared.Jogador{
		ID:       novoID,
		PosX:     5,
		PosY:     10,
		Elemento: shared.Elemento{Simbolo: '☺'},
	}

	js.estado.Jogadores[novoID] = novoJogador
	js.proximoID++

	reply.JogadorID = novoID
	reply.Estado = js.estado
	log.Printf("Jogador %d entrou", novoID)
	return nil
}

func (js *JogoServidor) GetEstadoJogo(args *struct{}, reply *shared.EstadoJogo) error {
	js.mutex.Lock()
	defer js.mutex.Unlock()
	*reply = js.estado
	return nil
}

func (js *JogoServidor) MoverJogador(args *shared.MoverRPC, reply *struct{}) error {
	js.mutex.Lock()
	defer js.mutex.Unlock()

	jogador, ok := js.estado.Jogadores[args.JogadorID]
	if !ok {
		return nil // jogador nao existe mais, ignora
	}

	// logica pra nao processar o mesmo comando duas vezes
	if args.SequenceNumber <= jogador.UltimoSequenceNumber {
		return nil
	}
	jogador.UltimoSequenceNumber = args.SequenceNumber

	dx, dy := 0, 0
	switch args.Tecla {
	case 'w':
		dy = -1
	case 'a':
		dx = -1
	case 's':
		dy = 1
	case 'd':
		dx = 1
	}
	novaPosX, novaPosY := jogador.PosX+dx, jogador.PosY+dy

	movimentoValido := true
	// checa se ta dentro do mapa e se nao eh parede
	if novaPosY < 0 || novaPosY >= len(js.estado.Mapa) || novaPosX < 0 || novaPosX >= len(js.estado.Mapa[novaPosY]) || js.estado.Mapa[novaPosY][novaPosX].Tangivel {
		movimentoValido = false
	}

	// checa se nao bateu em outro jogador
	if movimentoValido {
		for _, outroJogador := range js.estado.Jogadores {
			if outroJogador.ID != jogador.ID && novaPosX == outroJogador.PosX && novaPosY == outroJogador.PosY {
				movimentoValido = false
				break
			}
		}
	}

	if movimentoValido {
		jogador.PosX = novaPosX
		jogador.PosY = novaPosY
	}

	js.estado.Jogadores[args.JogadorID] = jogador
	return nil
}

func main() {
	jogoServidor := NovoJogoServidor()
	rpc.Register(jogoServidor)
	porta := ":1234"
	listener, err := net.Listen("tcp", porta)
	if err != nil {
		log.Fatal("Erro na porta: ", err)
	}
	defer listener.Close()

	log.Printf("Servidor escutando na porta %s", porta)
	rpc.Accept(listener)
}

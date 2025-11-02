package main

import (
	"bufio"
	"fmt"
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

func (js *JogoServidor) jogoCarregarMapaLadoServidor(caminhoArquivo string) error {
	arq, err := os.Open(caminhoArquivo)
	if err != nil {
		return err
	}
	defer arq.Close()
	scanner := bufio.NewScanner(arq)
	y := 0
	for scanner.Scan() {
		linha := scanner.Text()
		var linhaElems []shared.Elemento
		for _, ch := range linha {
			var e shared.Elemento
			switch ch {
			case '▤':
				e = shared.Elemento{Simbolo: ch, Cor: 8, CorFundo: 0, Tangivel: true}
			case '♣':
				e = shared.Elemento{Simbolo: ch, Cor: 2, CorFundo: 0, Tangivel: false}
			default:
				e = shared.Elemento{Simbolo: ' ', Cor: 0, CorFundo: 0, Tangivel: false}
			}
			linhaElems = append(linhaElems, e)
		}
		js.estado.Mapa = append(js.estado.Mapa, linhaElems)
		y++
	}
	return scanner.Err()
}

func NovoJogoServidor() *JogoServidor {
	js := &JogoServidor{proximoID: 1}
	js.estado.Jogadores = make(map[int]shared.Jogador)
	if err := js.jogoCarregarMapaLadoServidor("game/mapa.txt"); err != nil {
		log.Fatalf("Falha ao carregar o mapa: %v", err)
	}
	log.Println("Mapa do jogo carregado com sucesso.")
	return js
}

func (js *JogoServidor) RegistrarNovoJogador(args *struct{}, reply *shared.RespostaServidorRPC) error {
	js.mutex.Lock()
	defer js.mutex.Unlock()
	novoID := js.proximoID
	posX, posY := 5, 10
	novoJogador := shared.Jogador{
		ID:             novoID,
		PosX:           posX,
		PosY:           posY,
		Elemento:       shared.Elemento{Simbolo: '☺', Cor: 7},
		UltimoVisitado: js.estado.Mapa[posY][posX],
	}
	js.estado.Jogadores[novoID] = novoJogador
	js.proximoID++
	reply.JogadorID = novoID
	reply.Estado = js.estado
	log.Printf("Jogador %d registrado em (%d, %d). Total: %d", novoID, posX, posY, len(js.estado.Jogadores))
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
		return fmt.Errorf("jogador com ID %d não encontrado", args.JogadorID)
	}

	if args.SequenceNumber <= jogador.UltimoSequenceNumber {
		log.Printf("Comando duplicado do jogador %d. Seq: %d, último: %d. Ignorando.", args.JogadorID, args.SequenceNumber, jogador.UltimoSequenceNumber)
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
	if novaPosY < 0 || novaPosY >= len(js.estado.Mapa) || novaPosX < 0 || novaPosX >= len(js.estado.Mapa[novaPosY]) || js.estado.Mapa[novaPosY][novaPosX].Tangivel {
		movimentoValido = false
	}
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
		log.Printf("Jogador %d moveu-se para (%d, %d) [Seq: %d]", args.JogadorID, novaPosX, novaPosY, args.SequenceNumber)
	}

	js.estado.Jogadores[args.JogadorID] = jogador
	return nil
}

func main() {
	jogoServidor := NovoJogoServidor()
	rpc.Register(jogoServidor)
	log.Println("Servidor de Jogo RPC registrado.")
	porta := ":1234"
	listener, err := net.Listen("tcp", porta)
	if err != nil {
		log.Fatal("Erro ao escutar na porta: ", err)
	}
	defer listener.Close()
	log.Printf("Servidor escutando na porta %s", porta)
	fmt.Println("Pressione Ctrl+C para encerrar.")
	rpc.Accept(listener)
}

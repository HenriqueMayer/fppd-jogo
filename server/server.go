// Define a struct que guardará o estado do jogo.
// Define um método "placeholder" que será exposto via RPC.
// Inicia um servidor RPC que escuta por conexões de rede em uma porta específica.

package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"sync"

	"jogo/shared" // Certifique-se que o nome do módulo em go.mod é "jogo"
)

// JogoServidor é a struct principal do nosso servidor.
// Ela vai guardar o estado do jogo e o mutex para controle de concorrência.
type JogoServidor struct {
	mutex     sync.Mutex
	estado    shared.EstadoJogo
	proximoID int
}

// jogoCarregarMapaLadoServidor: lê um arquivo e constrói o mapa do jogo.
// É quase idêntico ao original, mas adaptado para as structs do pacote 'shared'.
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
			// Define os elementos do mapa com base nos caracteres do arquivo.
			// As cores são representadas por inteiros simples.
			switch ch {
			case '▤':
				e = shared.Elemento{Simbolo: ch, Cor: 8, CorFundo: 0, Tangivel: true} // Cinza escuro sobre preto
			case '♣':
				e = shared.Elemento{Simbolo: ch, Cor: 2, CorFundo: 0, Tangivel: false} // Verde sobre preto
			// Ignoramos o personagem e inimigo do mapa, pois eles serão dinâmicos.
			// Qualquer outro caractere vira um espaço vazio.
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

// NovoJogoServidor cria e inicializa uma nova instância do nosso servidor.
func NovoJogoServidor() *JogoServidor {
	js := &JogoServidor{
		proximoID: 1, // O primeiro jogador terá ID 1
	}
	// Importante: mapas em Go precisam ser inicializados antes do uso.
	js.estado.Jogadores = make(map[int]shared.Jogador)

	// Carrega o mapa do jogo.
	// O caminho é relativo à raiz do projeto, de onde executamos o 'go run'.
	if err := js.jogoCarregarMapaLadoServidor("game/mapa.txt"); err != nil {
		// Se o mapa não puder ser carregado, o servidor não pode iniciar.
		log.Fatalf("Falha ao carregar o mapa: %v", err)
	}
	log.Println("Mapa do jogo carregado com sucesso.")

	return js
}

// RegistrarNovoJogador é o método RPC que um cliente chama para entrar no jogo.
func (js *JogoServidor) RegistrarNovoJogador(args *struct{}, reply *shared.RespostaServidorRPC) error {
	// 1. Trava o mutex! Essencial para proteger o estado de acessos concorrentes.
	js.mutex.Lock()
	// 2. 'defer' garante que o mutex será destravado no final da função, não importa o que aconteça.
	defer js.mutex.Unlock()

	// 3. Cria um novo jogador
	novoID := js.proximoID
	posX, posY := 5, 10 // Posição inicial fixa (poderia ser aleatória)

	novoJogador := shared.Jogador{
		ID:   novoID,
		PosX: posX,
		PosY: posY,
		Elemento: shared.Elemento{
			Simbolo: '☺',
			Cor:     7, // Cor branca
		},
		// O 'UltimoVisitado' será o elemento que está no mapa nessa posição inicial.
		UltimoVisitado: js.estado.Mapa[posY][posX],
	}

	// 4. Adiciona o novo jogador ao estado do jogo
	js.estado.Jogadores[novoID] = novoJogador
	js.proximoID++ // Incrementa o ID para o próximo jogador

	// 5. Preenche a resposta que será enviada de volta para o cliente
	reply.JogadorID = novoID
	reply.Estado = js.estado // Envia a "fotografia" completa e atual do jogo

	log.Printf("Jogador %d registrado com sucesso em (%d, %d).", novoID, novoJogador.PosX, novoJogador.PosY)
	log.Printf("Total de jogadores agora: %d", len(js.estado.Jogadores))

	// 6. Retorna 'nil' para indicar que a operação RPC foi um sucesso.
	return nil
}

func main() {
	// Cria uma nova instância do nosso servidor de jogo.
	jogoServidor := NovoJogoServidor()

	// Registra a instância do jogo no sistema RPC.
	// Isso torna os métodos exportados (com letra maiúscula) de JogoServidor
	// disponíveis para serem chamados remotamente.
	rpc.Register(jogoServidor)
	log.Println("Servidor de Jogo RPC registrado.")

	// Cria um "listener" de rede na porta TCP 1234.
	// O servidor vai "escutar" por conexões nesta porta.
	porta := ":1234"
	listener, err := net.Listen("tcp", porta)
	if err != nil {
		log.Fatal("Erro ao escutar na porta: ", err)
	}
	// Garante que o listener será fechado quando a função main terminar.
	defer listener.Close()
	log.Printf("Servidor escutando na porta %s", porta)

	fmt.Println("Pressione Ctrl+C para encerrar o servidor.")

	// Loop infinito para aceitar conexões de clientes.
	// A função rpc.Accept é bloqueante e lida com cada cliente
	// em uma nova goroutine automaticamente.
	rpc.Accept(listener)
}

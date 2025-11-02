package main

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"jogo/shared"
)

type JogoCliente struct {
	mutex          sync.Mutex
	meuID          int
	estado         shared.EstadoJogo
	clienteRPC     *rpc.Client
	sequenceNumber int64
}

// -- Funcoes RPC do Cliente --

func (jc *JogoCliente) conectarAoServidor(endereco string) {
	cliente, err := rpc.Dial("tcp", endereco)
	if err != nil {
		log.Fatalf("Erro ao conectar: %v", err)
	}
	jc.clienteRPC = cliente
}

func (jc *JogoCliente) registrarNoServidor() {
	var resposta shared.RespostaServidorRPC
	err := jc.clienteRPC.Call("JogoServidor.RegistrarNovoJogador", &struct{}{}, &resposta)
	if err != nil {
		log.Fatalf("Erro ao registrar: %v", err)
	}
	jc.meuID = resposta.JogadorID
	jc.estado = resposta.Estado
	log.Printf("Entrei no jogo com ID %d", jc.meuID)
}

func (jc *JogoCliente) mover(tecla rune) {
	jc.mutex.Lock()
	jc.sequenceNumber++
	seq := jc.sequenceNumber
	jc.mutex.Unlock()

	args := &shared.MoverRPC{
		JogadorID:      jc.meuID,
		Tecla:          tecla,
		SequenceNumber: seq,
	}

	// tenta mandar o comando 3 vezes se der erro
	for i := 0; i < 3; i++ {
		var reply struct{}
		err := jc.clienteRPC.Call("JogoServidor.MoverJogador", args, &reply)
		if err == nil {
			return // sucesso
		}
		log.Printf("Erro ao mover, tentando de novo: %v", err)
		time.Sleep(200 * time.Millisecond)
	}
	log.Printf("Nao consegui mandar o comando de movimento")
}

// -- Goroutines --

// fica buscando o estado do jogo do servidor
func (jc *JogoCliente) sincronizarComServidor() {
	for {
		var estadoAtualizado shared.EstadoJogo
		err := jc.clienteRPC.Call("JogoServidor.GetEstadoJogo", &struct{}{}, &estadoAtualizado)
		if err != nil {
			log.Printf("Perdi conexao com servidor: %v", err)
			os.Exit(1)
		}

		jc.mutex.Lock()
		jc.estado = estadoAtualizado
		jc.mutex.Unlock()

		time.Sleep(100 * time.Millisecond)
	}
}

// le o teclado e manda as teclas por um canal
func lerEntradaDoTeclado(ch chan<- rune) {
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
	defer exec.Command("stty", "-F", "/dev/tty", "echo").Run()

	var b = make([]byte, 1)
	for {
		os.Stdin.Read(b)
		ch <- rune(b[0])
	}
}

// -- Logica de Desenho --

func (jc *JogoCliente) desenharEstado() {
	var sb strings.Builder
	sb.WriteString("\033[H\033[2J") // limpa a tela

	jc.mutex.Lock()
	mapa := jc.estado.Mapa
	jogadores := jc.estado.Jogadores
	meuID := jc.meuID
	jc.mutex.Unlock()

	if len(mapa) == 0 {
		sb.WriteString("Aguardando mapa...")
		return
	}

	mapaVisual := make([][]rune, len(mapa))
	for i, linha := range mapa {
		mapaVisual[i] = make([]rune, len(linha))
		for j, elem := range linha {
			mapaVisual[i][j] = elem.Simbolo
		}
	}

	for _, jogador := range jogadores {
		if jogador.PosY >= 0 && jogador.PosY < len(mapaVisual) && jogador.PosX >= 0 && jogador.PosX < len(mapaVisual[jogador.PosY]) {
			mapaVisual[jogador.PosY][jogador.PosX] = jogador.Elemento.Simbolo
		}
	}

	for _, linha := range mapaVisual {
		sb.WriteString(string(linha))
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("\nVoce e o Jogador %d\n", meuID))
	sb.WriteString("Use WASD para mover, 'q' para sair\n")
	fmt.Print(sb.String())
}

func main() {
	jc := &JogoCliente{}

	jc.conectarAoServidor("localhost:1234")
	defer jc.clienteRPC.Close()
	jc.registrarNoServidor()

	go jc.sincronizarComServidor()

	canalTeclado := make(chan rune)
	go lerEntradaDoTeclado(canalTeclado)

	// loop principal que desenha a tela e processa eventos
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		jc.desenharEstado()

		select {
		case tecla := <-canalTeclado:
			if tecla == 'q' {
				return
			}
			if tecla == 'w' || tecla == 'a' || tecla == 's' || tecla == 'd' {
				go jc.mover(tecla)
			}
		case <-ticker.C:
			// nao faz nada, so serve pra forcar o redesenho
		}
	}
}

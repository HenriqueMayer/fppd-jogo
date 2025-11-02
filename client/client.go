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

func (jc *JogoCliente) conectarAoServidor(endereco string) {
	log.Println("Tentando conectar ao servidor em", endereco, "...")
	cliente, err := rpc.Dial("tcp", endereco)
	if err != nil {
		log.Fatalf("FALHA CRÍTICA: Erro ao conectar ao servidor: %v", err)
	}
	jc.clienteRPC = cliente
	log.Println("SUCESSO: Conectado ao servidor RPC.")
}

func (jc *JogoCliente) registrarNoServidor() {
	log.Println("Tentando se registrar no servidor...")
	var resposta shared.RespostaServidorRPC
	err := jc.clienteRPC.Call("JogoServidor.RegistrarNovoJogador", &struct{}{}, &resposta)
	if err != nil {
		log.Fatalf("FALHA CRÍTICA: Erro ao registrar no servidor: %v", err)
	}
	jc.meuID = resposta.JogadorID
	jc.estado = resposta.Estado
	log.Printf("SUCESSO: Registrado! Meu ID é %d.", jc.meuID)
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

	maxTentativas := 3
	esperaEntreTentativas := 200 * time.Millisecond

	for tentativa := 1; tentativa <= maxTentativas; tentativa++ {
		if jc.clienteRPC == nil {
			return
		}
		var reply struct{}
		err := jc.clienteRPC.Call("JogoServidor.MoverJogador", args, &reply)
		if err == nil {
			return
		}

		log.Printf("ERRO: Falha ao enviar comando [Seq: %d] (tentativa %d de %d): %v", seq, tentativa, maxTentativas, err)
		if tentativa < maxTentativas {
			time.Sleep(esperaEntreTentativas)
		}
	}
	log.Printf("CRÍTICO: Não foi possível enviar o comando [Seq: %d] após %d tentativas.", seq, maxTentativas)
}

func (jc *JogoCliente) sincronizarComServidor() {
	log.Println("Goroutine de sincronização iniciada.")
	for {
		var estadoAtualizado shared.EstadoJogo
		if jc.clienteRPC == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		err := jc.clienteRPC.Call("JogoServidor.GetEstadoJogo", &struct{}{}, &estadoAtualizado)
		if err != nil {
			log.Printf("FALHA CRÍTICA: Conexão com o servidor perdida: %v", err)
			os.Exit(1)
		}

		jc.mutex.Lock()
		jc.estado = estadoAtualizado
		jc.mutex.Unlock()

		time.Sleep(100 * time.Millisecond)
	}
}

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

func (jc *JogoCliente) desenharEstado() {
	var sb strings.Builder
	sb.WriteString("\033[H\033[2J")

	jc.mutex.Lock()
	mapa := jc.estado.Mapa
	jogadores := jc.estado.Jogadores
	meuID := jc.meuID
	jc.mutex.Unlock()

	if len(mapa) == 0 {
		sb.WriteString("Aguardando mapa do servidor...")
	} else {
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
	}

	sb.WriteString(fmt.Sprintf("\nVocê é o Jogador %d. Posições:\n", meuID))
	for id, p := range jogadores {
		sb.WriteString(fmt.Sprintf(" - Jogador %d: (%d, %d)\n", id, p.PosX, p.PosY))
	}
	sb.WriteString("\nUse WASD para mover. 'q' para sair.\n")
	fmt.Print(sb.String())
}

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	log.Println("Cliente iniciando...")
	jc := &JogoCliente{}
	jc.conectarAoServidor("localhost:1234")
	defer jc.clienteRPC.Close()
	jc.registrarNoServidor()
	go jc.sincronizarComServidor()
	canalTeclado := make(chan rune)
	go lerEntradaDoTeclado(canalTeclado)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		jc.desenharEstado()
		select {
		case tecla := <-canalTeclado:
			if tecla == 'q' || tecla == 'Q' {
				return
			}
			if tecla == 'w' || tecla == 'a' || tecla == 's' || tecla == 'd' {
				go jc.mover(tecla)
			}
		case <-ticker.C: // Tick
		}
	}
}

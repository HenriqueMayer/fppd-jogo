> Obs.: Sempre que eu clico "ctrl+s" para salvar o código é formatado de forma padrão, por isso a formatação beme struturada.
!!
---
---

### Tópico 1

- Heterogeneidade: lidar com diferentes hardwares, sistemas operacionais e redes. O **Middleware é a solução!** (ele cria uma camada de abstração)

- Escalabilidade: o sistema deve continuar funcionando com 10 e com 10,000 jogadores. **Evitar CENTRALIZAÇÃO!**

- Tratamento de Falhas: o sistema deve continuar funcionando mesmo que partes dele falhem. Isso se consegue com **REDUNDÂNCIA, DETECÇÃO DE FALHAS e RECUPERAÇÃO**

- Concorrência: vários clientes acessarão o estado do jogo no servidor ao mesmo tempo. Precisamos garantir que eles não corrompam os dados. A ferramenta para isso é o **Mutex - Mutal Exclusion**, que garante que apenas um cliente possa modificar os dados por vez.

- Transparência: o objetivo dos SDs é esconder a complexidade.
    - Transparência de Acesso e Localização: o cliente não precisa saber o endereço IP do servidor, ele pode usar um nome qualquer.
    - Transparência de Falhas: o cliente deve tentar se reconectar automaticamente se o servidor cair sem o usuário fazer nada.

 
### Tópico 2
- Comunicação entre Processos - IPC

| Como as máquinas em um SD conversam?

1. Memória Compartilhada: é rápido, mas só funciona na mesma máquina. Então está descartado.
2. **Troca de Mensagens**: é a base da comunicação em rede. O cliente "empacota" os dados (ex.: quero mover para a direita) em uma mensagem e a envia para o servidor.

- Troca de Mensagem
   - Sincronização:
        1. Síncrona (*Bloqueante*): o cliente envia uma mes=nsagem e **espera** pela resposta antes de continuar.
        2. Assíncrona (*Não-Bloqueante*): o cliente envia a mensagem e continua fazendo outras coisas. (Ex.: um e-mail)
    
    - Codificação/Decodificação:
        1. Um computador pode usar o padrão Big-Endian e outro Little-Endian para armazenar números. Para que eles se entendam, eles precisam concordar com um idioma para a rede, uma sintaxe de transferência (JSON, Protocol Buffers, XDR)

    - Tratamento de Falhas e Idempotência:
        1. E se o cliente envia a mensagem "atirar", mas a resposta do servidor se perde na rede? Ele pode enviar novamente a mensagem.
        2. Mas se o servidor receber a mensagem duas vezes ele pode atirar duas vezes, isso é ruim!
        3. Uma operação idempotente é aquela que, se repetida, produz o mesmo resultado (setVida(100))
        4. Operações não-idempotentes (atirar(), coletar()) precisam de uma semântica de **"exactly-once"**, que é um dos requisitos mais importantes.

- Remote Procedure Call - RPC
    - Objetivo: fazer uma chamada de função/método em outra máquina parecer exatamente como uma chamada de função local.

| Como?

1. **Chamada do Cliente:** você chama a função *Multiplicar(5, 10)* que, na verdade, não existe no seu código. O que existe é um objeto chamado **Stub do Cliente**
2. **Stub do Cliente:** esse código, geralmente gerado automaticamente, tem a mesma assinatura da função remota. Ele pega seus argumentos (5 e 10), empacota-os em uma mensagem (marshalling) e envia pela rede para o servidor. A sua aplicação fica **bloqueada**, esperando a resposta.
3. **Rede:** a mensagem viaja até o servidor.
4. **Stub do Servidor:** no lado do servidor, um Stub recebe a mensagem, desempacota os argumentos (unmarshalling) e chama a função real *Multiplicar(5, 10).
5. **Execução do Servidor:** a função real executa e retorna **50**.
6. **Retorno:** o Stub do Servidor pega o resultado, empacota o 50 e o envia pela rede.
7. **Resposta ao Cliente:**  o Stub do Cliente recebe a resposta, desempacota o 50 e o retorna para sua aplicação.
8. **Continuação:** sua aplicação é desbloqueada e a variável resultado agora contém 50.

**RPC em Go é net/rpc**

- Deve ser um método de um tipo de dado exportado (nome começa com maiúscula)
- Deve ter dois argumentos, amobos do tipo exportado
- O segundo argumento deve ser um ponteiro (é onde a resposta será colocada)
- Deve retornar um valor do tipo <error>

Exemplo: <func (t *MeuTipo) MeuMetodo(args *Argumentos, reply *Resposta) error>

### Aula de Go

```Go
// Todo arquivo começa com a declaração do apcote.
// "main" é especial: define um programa executável.
package main

import "fmt"

func main () {
    fmt.Println("Hello, World!")
}
```

--------------------------------------------------------

```Go
// Declaração explícita
var idade int=30

// Inferência de tipo com o operado "marmota" := (apenas dentro de funções)
nome := "Marcelo"
vivo := true

//Structs: uma fomra de agrupar dados:
tpe Jogador struct{
    ID int
    PosX int
    PosY int
}

// Criando uma instância de uma struct
jogador1 := Jogador{ID: 1, PosX: 10, PosY: 5}
```

--------------------------------------------------------

```Go
import "errors"

func dividir(a, b float64) (float64, error) {
    if b==0 {
        return 0, errors.New("não foi possível")
    }
    return a/b, nil
}

func main() {
    resultado, err := dividir(10,2)
    if err != nil {
        fmt.Println("Ocorreu um erro:", err)
    } else {
        fmt.Println("Resultado:", resultado)
    }
}
```

--------------------------------------------------------

```Go
func tarefaDemorada() {
    <...>
    fmt.Println("Terminei")
}

func main() {
    go tarefaDemorada()

    fmt.Println("A main terminou, mas a goroutine pode ainda estar rodando")
}
```

--------------------------------------------------------

- No nosso cliente, usaremos uma goroutine para buscar atualizações do servidor em segundo plano, enquanto a *main* cuida da entrada do teclado.
- *sync.Mutex* é a ferramente para resolver o problema da concorrência. Quando múltiplas goroutines (ou chamadas de clientes) tentam modificar o mesmo dado ao mesmo tempo temos uma condição de corrida. O mutex portege os dados.

```Go
import (
    "fmt"
    "sync"
)

// Estado do jogo que será compartilhado
var estadoDoJogo = make(map[string]int)

// O mutex que vai proteger o nosso estado
func escreverNoEstado(id string) {
    // Bloqueia o acesso. QUaluqer outra goroutine que tentar chamar Lock() agora ficará esperando
    mutex.Lock()

    // Acesso seguro para os dados compartilhados
    estadoDoJogo[id]++

    // Libera o bloqueio, permitindo que outra goroutine entre.
    mutex.Unlock()
}

func main() {
    // Simulação de 1000 clientes
    for i := 0; i < 1000; i++ {
        go escreverNoEstado("contador")
    }
    // Para ver o resultado precisamos esperar as goroutines terminarem
}

```
- No nosso servidor, a **struct** que contém o estado do jogo (lista de jogadores, posições, etc) 
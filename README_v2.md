### Documentando o processo

1. Primeiro fiz a implementação do shared, ele é como se fosse um contrato, onde o servidor e o cliente podem conversar no mesmo idioma com a struct bem definida.
2. Primeiros testes com o servidor:
    - Primeiro error: server/server.go:13:2: package fppd-jogo/shared is not in std (/usr/lib/go-1.22/src/fppd-jogo/shared)
    - Solução: eu precisava usar o nome do mod init que fiz a primeira vez, que no caso era apenas "jogo"
3. O servidor agora carrega um mapa estático ao iniciar, dando vida ao nosso mundo, método RPC RegistrarNovoJogador e mutex.
    - Problema:
    ```bash
    henriquemayer@hm-IdeaPad:~/Documents/college/fourth-semester/fppd-jogo$ go run ./server
    2025/11/02 19:08:58 Mapa do jogo carregado.
    2025/11/02 19:08:58 Servidor de Jogo RPC registrado.
    2025/11/02 19:08:58 Erro ao escutar na porta: listen tcp :1234: bind: address already in use
    exit status 1
    ```
    - Isso indica que tem uma instância anterior rodando em segundo plano.
    - Vou precisar dar kill no processo:
    ```bash
    henriquemayer@hm-IdeaPad:~$ sudo lsof -i :1234
    [sudo] password for henriquemayer:     
    COMMAND   PID          USER   FD   TYPE DEVICE SIZE/OFF NODE NAME
    server  53875 henriquemayer    3u  IPv6 192899      0t0  TCP *:1234 (LISTEN)

    kill 53875

    kill -9 53875
    ```
    -  Tive que usar o "kill -9 53875" para funcionar
    - Funcionou:
    ```bash
    henriquemayer@hm-IdeaPad:~/Documents/college/fourth-semester/fppd-jogo$ go run ./server
    2025/11/02 19:16:48 Mapa do jogo carregado.
    2025/11/02 19:16:48 Servidor de Jogo RPC registrado.
    2025/11/02 19:16:48 Servidor escutando na porta :1234
    ```
4. Implementando o Client.
5. panic: nil pointer dereference
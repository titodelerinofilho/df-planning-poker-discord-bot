# DF Planning Poker Discord Bot

Bot de Planning Poker para Discord escrito em Go.

## Desenvolvimento

Este repositório segue um fluxo de uma issue por vez, monólito modular e regras de domínio isoladas de Discord e PostgreSQL.

Comandos principais:

```bash
make run
make test
make test-race
make vet
make lint
make build
```

Os comandos usam `GOCACHE=.cache/go-build` por padrão para manter o cache local ao workspace.

Para executar `make lint`, instale o `golangci-lint` seguindo a documentação oficial do projeto. A configuração local fica em `.golangci.yml` e usa o formato v2 do `golangci-lint`.

Esta configuração foi validada com `golangci-lint v2.8.0`, compatível com o Go 1.24 definido no módulo. Caso o binário esteja fora do `PATH`, use:

```bash
GOLANGCI_LINT=/caminho/para/golangci-lint make lint
```

Exceções de lint documentadas: nenhuma regra específica do projeto foi desabilitada nesta fase.

## CI

O GitHub Actions executa build, testes, race detector, `go vet` e `golangci-lint` em pull requests e pushes para `main`. O workflow usa a versão de Go declarada em `go.mod` e cache automático do `actions/setup-go`.

## Docker

A imagem do bot usa build multi-stage com Go 1.24 e runtime distroless não root.

```bash
make docker-build
```

O container espera as mesmas variáveis obrigatórias descritas na seção de configuração. Ainda não há `HEALTHCHECK` no Dockerfile porque o bot não expõe endpoint HTTP nem comando de probe nesta fase.

## Compose Local

O ambiente local com Docker Compose sobe o bot e um PostgreSQL com volume persistente e healthcheck via `pg_isready`.

```bash
docker compose up --build
docker compose down
```

O Compose usa valores padrão seguros para desenvolvimento local. Para usar credenciais reais do Discord, exporte as variáveis de ambiente antes de subir os serviços ou crie um `.env` local não versionado.

## Configuração

Copie `.env.example` para um arquivo local não versionado e preencha os valores reais no ambiente antes de iniciar o bot.

Variáveis obrigatórias:

- `APP_ENV`: ambiente da aplicação, como `development`.
- `LOG_LEVEL`: nível de log estruturado, aceitando `debug`, `info`, `warn` ou `error`.
- `DISCORD_TOKEN`: token do bot do Discord. Nunca commite o valor real.
- `DISCORD_APPLICATION_ID`: ID da aplicação Discord.
- `DISCORD_GUILD_ID`: ID da guild de desenvolvimento. Obrigatório quando `COMMAND_REGISTRATION_MODE=guild`.
- `DATABASE_URL`: URL de conexão PostgreSQL. Nunca commite credenciais reais.
- `COMMAND_REGISTRATION_MODE`: `guild` para desenvolvimento ou `global` para produção.
- `SESSION_EXPIRATION`: duração de expiração das sessões, como `24h`.
- `SHUTDOWN_TIMEOUT`: duração máxima de encerramento, como `10s`.

O startup falha com uma mensagem clara quando uma variável obrigatória está ausente ou inválida, sem imprimir secrets.

## Discord

O bot inicializa uma sessão `discordgo` no startup usando `DISCORD_TOKEN` e configura somente o intent mínimo de guilds nesta fase.

Quando `COMMAND_REGISTRATION_MODE=guild`, o startup sincroniza comandos de aplicação na guild informada por `DISCORD_GUILD_ID`. Use este modo em desenvolvimento porque comandos de guild propagam rapidamente.

Quando `COMMAND_REGISTRATION_MODE=global`, o startup sincroniza comandos globais da aplicação, sem usar `DISCORD_GUILD_ID`. Use este modo em produção e considere que o Discord pode levar vários minutos para propagar alterações de comandos globais.

A sincronização roda apenas no startup: ela cria comandos ausentes, atualiza definições alteradas e remove somente comandos com nomes explicitamente gerenciados pelo bot. Interações recebidas não registram comandos.

O comando `/ping` responde de forma efêmera com a latência básica do gateway.

Componentes interativos usam `custom_id` no formato `namespace:action:resource`. O recurso deve ser um identificador opaco e não deve carregar voto, token ou informação sensível.

Mensagens apresentadas pelo bot usam controle restritivo de menções por padrão, evitando disparos acidentais de `@everyone`, cargos ou usuários a partir de conteúdo livre.

Os logs são emitidos em JSON via `slog` e incluem os campos fixos `version`, `environment`, `channel` e `correlation_identifier`.

Os canais seguem esta separação:

- `exceptions`: STDERR, nível `CRITICAL`.
- `requests`: STDOUT, nível `INFO`.
- `responses`: STDOUT, nível `INFO`.
- `database_queries`: STDOUT, nível `INFO`.

## Domínio

A escala padrão de Planning Poker é Fibonacci modificada: `0, 1, 2, 3, 5, 8, 13, 21, 34, 55, ?, ☕`. Os valores `?` e `☕` são especiais e não representam números.

As estatísticas numéricas calculam mínimo, máximo, mediana e moda usando somente estimativas numéricas. Rodadas com apenas valores especiais são representadas sem resultado numérico.

Sessões de Planning Poker nascem no estado `JOINING`, vinculadas a guild, canal, thread, mensagem principal, tarefa, criador, facilitador, escala e expiração. Participantes e rodadas são adicionados nas próximas etapas da máquina de estados.

Enquanto a sessão está em `JOINING`, participantes podem entrar e sair. O domínio impede duplicidade de participantes ativos e bloqueia mudanças na lista depois que a sessão sai desse estado.

Ao fechar a lista de participantes, a sessão exige a quantidade mínima configurada de participantes ativos, muda para `VOTING` e cria a primeira rodada em estado `OPEN`.

Durante `VOTING`, somente participantes ativos podem votar usando valores da escala da sessão. O participante pode alterar o voto antes da revelação, e quando todos os participantes ativos votam a sessão muda para `READY_TO_REVEAL`.

## Encerramento

O processo trata `SIGINT` e `SIGTERM`, cancela o contexto raiz e executa o fechamento de recursos respeitando `SHUTDOWN_TIMEOUT`.

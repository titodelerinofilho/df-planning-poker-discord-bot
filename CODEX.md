# CODEX.md — Regras para implementação

## Contexto

Você está trabalhando no projeto **Planning Poker Discord Bot**, um bot de estimativas colaborativas para times de desenvolvimento.

O projeto é escrito em Go e utiliza a API do Discord por meio da biblioteca `discordgo`. PostgreSQL é a fonte de verdade. O bot usa comandos slash, componentes interativos, respostas efêmeras e threads para organizar sessões de Planning Poker.

Uma sessão recebe o link de uma issue ou tarefa. Os participantes entram na sessão, votam de forma secreta e podem alterar o voto até a revelação. Quando todos votarem, o bot informa que a rodada está pronta. Um facilitador autorizado revela os votos. Quando houver divergência relevante, o bot menciona os participantes dos extremos ou votos discrepantes e solicita contexto. A sessão pode ter novas rodadas e termina com uma estimativa acordada.

Este repositório começa como monólito modular. Não transforme o sistema em microserviços sem uma decisão arquitetural explícita.

## Regra principal

Trabalhe em **uma issue por vez**. Não tente implementar um épico inteiro em um prompt.

Antes de escrever código:

1. leia integralmente a issue;
2. leia `AGENTS.md`, `CODEX.md`, `ROADMAP.md`, `README.md` e documentos relacionados;
3. identifique o comportamento esperado;
4. identifique módulos afetados;
5. liste suposições;
6. apresente um plano curto;
7. implemente somente o escopo solicitado;
8. organize a mudança em commits lógicos;
9. rode formatação, análise estática e testes;
10. atualize a documentação;
11. apresente um resumo final com arquivos alterados, testes e pendências.

Não implemente requisitos futuros do roadmap apenas porque parecem úteis.

## Objetivo do produto

O bot deve permitir o seguinte fluxo:

1. um facilitador inicia uma sessão com o link da tarefa;
2. o bot cria uma thread e uma mensagem principal;
3. usuários entram ou são adicionados como participantes;
4. a lista de participantes é fechada;
5. cada participante vota secretamente;
6. a mensagem pública mostra apenas quem já votou;
7. o bot sinaliza quando todos votaram;
8. um facilitador revela a rodada;
9. o bot mostra votos e estatísticas;
10. divergências são destacadas de forma respeitosa;
11. o time pode discutir e iniciar nova rodada;
12. o facilitador registra a estimativa final.

## Restrições arquiteturais

- domínio não importa `discordgo`, PostgreSQL, HTTP ou SDK de provedor;
- PostgreSQL é a fonte de verdade;
- nenhuma sessão aberta pode existir somente em memória;
- integrações externas são adaptadores;
- regras de sessão, rodada, participante, voto, escala e divergência ficam no domínio;
- handlers do Discord não contêm regra de negócio;
- presenters do Discord não decidem transições de estado;
- toda mutação relevante usa transação;
- todo I/O recebe `context.Context`;
- toda integração tem timeout;
- concorrência é limitada;
- callbacks e goroutines respeitam cancelamento;
- nenhum voto secreto aparece em logs;
- nenhum token ou secret entra no Git;
- IDs de componentes são opacos e não carregam o voto;
- autorização é validada no backend;
- interações repetidas devem ser idempotentes quando possível;
- migrations devem ser seguras para deploy progressivo;
- modelos do Discord e provedores externos não vazam para o domínio;
- não criar abstração sem caso real;
- não adicionar Redis, Kafka ou microserviços sem issue e justificativa explícitas.

## Dependências recomendadas

Use a biblioteca padrão sempre que ela for suficiente.

Dependências previstas:

```text
github.com/bwmarrin/discordgo
github.com/jackc/pgx/v5
github.com/google/uuid ou gerador equivalente
github.com/pressly/goose/v3, golang-migrate ou ferramenta escolhida pelo projeto
github.com/stretchr/testify apenas se já adotado
golang.org/x/sync/errgroup
```

Não introduza ORM completo sem decisão explícita.

A ferramenta de migrations deve ser única no projeto.

## Padrões Go

- código simples;
- dependências explícitas;
- interfaces pequenas e definidas pelo consumidor;
- construtores;
- invariantes protegidas no domínio;
- erros com wrapping;
- códigos e erros estáveis;
- `slog`;
- `context`;
- graceful shutdown;
- `errgroup` para concorrência estruturada;
- worker pools limitados;
- testes table-driven;
- race detector;
- evitar globals mutáveis;
- clock e ID generator injetáveis;
- não criar abstração sem caso real;
- não usar goroutine sem estratégia de cancelamento e tratamento de erro;
- não usar `panic` para erro operacional;
- evitar `init`;
- evitar dependências circulares;
- usar tipos próprios para IDs e estados quando isso proteger invariantes.

## Estilo e legibilidade

Preferir código com respiro visual entre preparação, execução, validação de erro e retorno.

Evitar declarações curtas dentro do cabeçalho de `if`, como:

```go
if err := do(); err != nil {
	return err
}
```

Preferir:

```go
err := do()

if err != nil {
	return fmt.Errorf("do operation: %w", err)
}
```

Para retorno com resultado:

```go
session, err := repository.FindByID(ctx, sessionID)

if err != nil {
	return Session{}, fmt.Errorf("find planning session: %w", err)
}

return session, nil
```

Outras regras:

- inserir linha em branco antes de blocos `if`;
- evitar funções longas com muitas responsabilidades;
- extrair helpers somente quando melhorarem a leitura;
- não compactar código apenas para reduzir linhas;
- priorizar clareza sobre concisão;
- usar nomes da linguagem do domínio;
- evitar `utils`, `helpers`, `common`, `manager` e `service` genéricos;
- comentários devem explicar intenção, risco ou decisão;
- não comentar código óbvio;
- não criar interface ao lado da implementação por padrão;
- definir interface no pacote consumidor.

## Estrutura de módulo

```text
cmd/
├── bot/
└── migrate/

internal/
├── domain/
│   ├── planning/
│   └── guild/
├── application/
│   ├── planning/
│   └── ports/
├── adapters/
│   ├── discord/
│   ├── postgres/
│   ├── github/
│   ├── gitlab/
│   └── jira/
├── platform/
│   ├── config/
│   ├── database/
│   ├── logging/
│   ├── telemetry/
│   └── shutdown/
└── testutil/
```

Não crie todos os diretórios antecipadamente. Crie-os conforme as issues exigirem.

## Domínio

### Agregado principal

`PlanningSession` é o agregado principal.

Ele controla:

- estado da sessão;
- rodada ativa;
- participantes;
- autorização de transições;
- abertura e fechamento de votação;
- revelação;
- reinício;
- conclusão;
- cancelamento;
- expiração.

Não permita que handlers alterem campos diretamente.

Prefira métodos expressivos:

```go
session.AddParticipant(...)
session.CloseParticipants(...)
session.CastVote(...)
session.Reveal(...)
session.StartNextRound(...)
session.Complete(...)
session.Cancel(...)
session.Expire(...)
```

### Estados

Sessão:

```text
JOINING
VOTING
READY_TO_REVEAL
REVEALED
DISCUSSING
COMPLETED
CANCELLED
EXPIRED
```

Rodada:

```text
OPEN
READY
REVEALED
CLOSED
```

Toda transição deve ser testada.

### Escala

A escala padrão é:

```text
0, 1, 2, 3, 5, 8, 13, 21, 34, 55, ?, ☕
```

A escala deve:

- validar votos;
- informar posição ordinal;
- separar valores numéricos e especiais;
- produzir representação segura para a camada de apresentação.

Não espalhe slices de valores por handlers.

### Divergência

A divergência deve ser calculada pela posição na escala.

Regra inicial:

- ignore `?` e `☕` nos cálculos;
- encontre a mediana ordinal;
- destaque votos com distância ordinal absoluta maior ou igual a dois;
- destaque extremos quando a amplitude ordinal exigir discussão;
- não rotule participantes como errados;
- gere dados estruturados; o texto final pertence ao presenter do Discord.

## Casos de uso

Cada arquivo de caso de uso deve representar uma ação clara.

Exemplos:

```text
StartSession
JoinSession
LeaveSession
CloseParticipants
CastVote
GetSessionStatus
RevealRound
RestartRound
CompleteSession
CancelSession
ExpireSessions
```

Um caso de uso deve:

1. validar entrada superficial;
2. carregar dados;
3. validar autorização;
4. chamar o domínio;
5. persistir em transação;
6. emitir resultado estruturado;
7. disparar efeitos externos de forma segura.

Não use estruturas de `discordgo` em inputs ou outputs da aplicação.

## Integração com Discord

A camada Discord é responsável por:

- registrar comandos;
- interpretar interações;
- responder dentro do prazo;
- adiar respostas quando necessário;
- criar e gerenciar threads;
- construir embeds e componentes;
- enviar respostas efêmeras;
- editar a mensagem principal;
- converter erros em mensagens úteis;
- validar o contexto básico de guild e canal.

A camada Discord não é responsável por:

- decidir se uma sessão pode revelar;
- calcular divergência;
- validar estimativa;
- decidir participantes pendentes;
- persistir regra de negócio.

### Comandos previstos

```text
/planning start
/planning status
/planning reveal
/planning revote
/planning finish
/planning cancel
/planning configure
```

Os nomes finais podem ser localizados ou ajustados por issue.

### Componentes previstos

```text
Join
Leave
Close participants
Vote
Reveal
Revote
Finish
Cancel
```

IDs devem identificar ação e recurso sem expor informação sensível.

Exemplo aceitável:

```text
planning:vote:<opaque-session-id>
```

Exemplo proibido:

```text
planning:vote:user-123:estimate-13
```

### Voto secreto

- confirmação de voto deve ser efêmera;
- mensagem pública exibe apenas estado de participação;
- logs não incluem valor;
- erros não incluem valor;
- persistência protege leitura antes da revelação pela camada de aplicação;
- somente resultado revelado pode ser apresentado publicamente.

## Persistência

Use PostgreSQL com `pgx`.

Repositórios devem retornar entidades ou modelos internos, não tipos de banco para a aplicação.

Tabelas previstas:

```text
planning_sessions
planning_rounds
planning_participants
planning_votes
guild_settings
integration_connections
audit_events
```

Constraints mínimas:

```text
UNIQUE(session_id, round_number)
UNIQUE(session_id, discord_user_id)
UNIQUE(round_id, discord_user_id)
```

Toda query multi-guild deve filtrar `guild_id` direta ou indiretamente por uma relação validada.

### Concorrência

Considere especialmente:

- dois votos simultâneos completando a rodada;
- alteração de voto concorrente com revelação;
- dois facilitadores revelando;
- clique repetido em botão;
- retry de interação;
- reinicialização do processo entre persistência e edição de mensagem.

Use transação e locking apropriados. Não resolva concorrência somente com mutex em memória.

## Efeitos externos e consistência

Mudança de domínio e atualização da mensagem do Discord não fazem parte da mesma transação distribuída.

Adote uma estratégia explícita:

1. persista o novo estado;
2. tente atualizar o Discord;
3. registre falha de sincronização;
4. permita reconciliação ou repetição idempotente.

Não reverta voto ou revelação apenas porque uma edição de mensagem falhou.

A mensagem principal deve poder ser reconstruída a partir do banco.

## Erros

Use wrapping com contexto.

Preferido:

```go
session, err := repository.FindOpenByThreadID(ctx, threadID)

if err != nil {
	return Output{}, fmt.Errorf("find open planning session by thread: %w", err)
}
```

Erros esperados devem ser identificáveis com `errors.Is`.

Não compare mensagens de erro.

Não exponha erro técnico bruto ao usuário do Discord.

## Configuração

Variáveis previstas:

```text
APP_ENV
LOG_LEVEL
DISCORD_TOKEN
DISCORD_APPLICATION_ID
DISCORD_GUILD_ID
DATABASE_URL
COMMAND_REGISTRATION_MODE
SESSION_EXPIRATION
SHUTDOWN_TIMEOUT
```

Regras:

- validar tudo no startup;
- não acessar `os.Getenv` fora do pacote de configuração;
- não fornecer default inseguro;
- não imprimir secrets;
- documentar `.env.example`;
- `DISCORD_GUILD_ID` pode ser obrigatório somente no modo de desenvolvimento.

## Logs

Use `slog` estruturado.

Nunca registrar:

- token;
- segredo de integração;
- valor de voto antes da revelação;
- payload integral de interação;
- conteúdo privado de issue sem necessidade.

Campos úteis:

```text
operation
guild_id
thread_id
session_id
round_id
actor_discord_id
interaction_id
duration_ms
result
error_code
```

## Testes

### Domínio

Cobrir:

- transições válidas;
- transições inválidas;
- entrada e saída;
- fechamento de participantes;
- voto;
- alteração de voto;
- conclusão automática da coleta;
- revelação;
- divergência;
- revotação;
- conclusão;
- cancelamento;
- expiração;
- valores especiais.

### Aplicação

Cobrir:

- autorização;
- transações;
- repositório não encontrado;
- idempotência;
- falha de dependência;
- cancelamento de contexto;
- efeitos externos.

### PostgreSQL

Usar banco real em testes de integração quando possível.

Cobrir:

- constraints;
- upsert de voto;
- locking;
- concorrência;
- migrations;
- consultas por guild, thread e sessão.

### Discord

Handlers devem ser testados com interações simuladas e portas falsas.

Não dependa de servidor real do Discord para a maioria dos testes.

## Comandos obrigatórios antes de concluir

Execute os comandos disponíveis no projeto, preferencialmente por `make`:

```bash
gofmt -w .
go test ./...
go test -race ./...
go vet ./...
```

Quando configurado:

```bash
golangci-lint run
```

Se algum comando não puder ser executado, informe claramente o motivo.

Nunca afirme que testes passaram sem executá-los.

## Migrations

- nunca altere migration já aplicada em ambiente compartilhado;
- crie uma nova migration;
- preserve compatibilidade de rollout;
- use UTC para instantes;
- use tipos adequados para IDs;
- adicione índices para consultas reais;
- avalie lock antes de adicionar constraint;
- inclua rollback quando seguro;
- não apague dados sem requisito explícito.

## Documentação

Atualize documentação quando alterar:

- comando;
- variável;
- migration;
- estado;
- fluxo;
- permissão;
- integração;
- regra de divergência;
- estratégia de deploy.

Documentação mínima:

```text
README.md
.env.example
docs/architecture.md
docs/discord-setup.md
docs/commands.md
docs/database.md
docs/deployment.md
```

## Commits

Use commits pequenos e coerentes.

Formato sugerido:

```text
feat(domain): add planning session state machine
feat(discord): handle secret vote interaction
feat(postgres): persist planning rounds
fix(voting): block vote after round reveal
test(application): cover duplicate reveal
docs: add Discord setup guide
```

Não inclua refatorações não relacionadas no mesmo commit.

## Resposta final do Codex

Ao finalizar uma issue, responda com:

### Implementado

- itens objetivos.

### Arquivos principais

- caminhos alterados e finalidade.

### Decisões

- decisões arquiteturais ou de domínio.

### Testes executados

```text
comando
resultado
```

### Migrations e configuração

- migrations criadas;
- variáveis novas;
- passos manuais.

### Pendências

- somente pendências reais;
- não sugerir funcionalidades aleatórias fora da issue.

## Critérios de recusa

Não implemente sem alertar quando a solicitação:

- expõe token;
- publica voto antes da revelação;
- remove autorização;
- armazena sessão apenas em memória;
- mistura domínio com `discordgo`;
- exige goroutine sem cancelamento;
- ignora transação em mutação crítica;
- inclui uma quantidade ampla de requisitos incompatível com uma única issue;
- contradiz uma regra explícita sem decisão documentada.

Nesses casos, explique o risco e proponha a menor alternativa segura compatível com a issue.

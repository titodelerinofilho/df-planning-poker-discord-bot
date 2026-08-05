# AGENTS.md — DF Planning Poker Discord Bot

## 1. Finalidade

Este documento orienta agentes de IA e colaboradores que trabalham neste repositório.

O projeto é um bot de Planning Poker para Discord, escrito em Go. O bot permite iniciar sessões de estimativa a partir de uma issue ou tarefa, reunir participantes em uma thread, registrar votos secretos, avisar quando todos votarem, revelar os resultados sob comando e conduzir discussões quando houver divergência relevante.

O objetivo principal é oferecer um fluxo simples, auditável e confiável para estimativas colaborativas dentro do Discord.

## 2. Regra de trabalho

Trabalhe em uma issue por vez.

Antes de alterar código:

1. leia integralmente a issue;
2. leia `README.md`, `AGENTS.md`, `CODEX.md` e documentos relacionados;
3. identifique o comportamento esperado;
4. identifique módulos afetados;
5. liste suposições e riscos;
6. apresente um plano curto;
7. implemente somente o escopo da issue;
8. rode formatação, análise estática e testes;
9. atualize a documentação afetada;
10. informe claramente o que foi alterado e o que ficou pendente.

Não implemente um épico completo quando a issue solicitar apenas uma parte.

## 3. Escopo funcional

O produto deve suportar, progressivamente:

- criação de sessão por comando slash;
- vínculo da sessão com uma URL de issue ou tarefa;
- criação ou reutilização de thread;
- entrada e saída de participantes;
- fechamento da lista de participantes;
- voto secreto;
- alteração de voto antes da revelação;
- acompanhamento de quem já votou, sem revelar valores;
- aviso quando todos os participantes votarem;
- revelação autorizada;
- cálculo de mínimo, máximo, mediana e moda;
- identificação de divergências;
- menção aos participantes com votos extremos ou discrepantes;
- nova rodada após discussão;
- finalização com estimativa acordada;
- cancelamento e expiração;
- histórico de sessões e rodadas;
- configuração por servidor;
- integrações opcionais com GitHub, GitLab e Jira.

## 4. Fora do escopo inicial

Não implementar no MVP, salvo issue explícita:

- microserviços;
- Kafka;
- Kubernetes;
- painel web;
- inteligência artificial para escolher a estimativa;
- atualização automática de issue externa;
- suporte a múltiplos bancos;
- cobrança ou planos;
- sistema complexo de plugins;
- abstrações genéricas para provedores que ainda não existem;
- Redis sem necessidade comprovada.

## 5. Princípios arquiteturais

O projeto deve começar como monólito modular.

A arquitetura deve separar:

- domínio;
- casos de uso;
- portas;
- adaptadores;
- inicialização e infraestrutura.

O domínio não conhece Discord, PostgreSQL, HTTP, SDKs externos ou variáveis de ambiente.

A aplicação orquestra casos de uso e depende de interfaces.

Os adaptadores implementam integração com Discord, banco, relógio, geração de IDs e provedores externos.

Dependências apontam para dentro:

```text
Discord / PostgreSQL / APIs externas
              │
              ▼
          Adapters
              │
              ▼
         Application
              │
              ▼
            Domain
```

## 6. Estrutura sugerida

```text
.
├── cmd/
│   ├── bot/
│   │   └── main.go
│   └── migrate/
│       └── main.go
├── internal/
│   ├── domain/
│   │   ├── planning/
│   │   │   ├── session.go
│   │   │   ├── round.go
│   │   │   ├── participant.go
│   │   │   ├── vote.go
│   │   │   ├── scale.go
│   │   │   ├── divergence.go
│   │   │   ├── errors.go
│   │   │   └── repository.go
│   │   └── guild/
│   │       ├── settings.go
│   │       └── repository.go
│   ├── application/
│   │   ├── planning/
│   │   │   ├── start_session.go
│   │   │   ├── join_session.go
│   │   │   ├── leave_session.go
│   │   │   ├── close_participants.go
│   │   │   ├── cast_vote.go
│   │   │   ├── reveal_round.go
│   │   │   ├── restart_round.go
│   │   │   ├── complete_session.go
│   │   │   ├── cancel_session.go
│   │   │   └── get_status.go
│   │   └── ports/
│   │       ├── transaction.go
│   │       ├── clock.go
│   │       ├── id_generator.go
│   │       ├── notifier.go
│   │       └── issue_provider.go
│   ├── adapters/
│   │   ├── discord/
│   │   │   ├── bot.go
│   │   │   ├── commands.go
│   │   │   ├── interactions.go
│   │   │   ├── components.go
│   │   │   ├── threads.go
│   │   │   ├── permissions.go
│   │   │   └── presenter.go
│   │   ├── postgres/
│   │   │   ├── session_repository.go
│   │   │   ├── guild_repository.go
│   │   │   └── transaction.go
│   │   ├── github/
│   │   ├── gitlab/
│   │   └── jira/
│   ├── platform/
│   │   ├── config/
│   │   ├── database/
│   │   ├── logging/
│   │   ├── telemetry/
│   │   └── shutdown/
│   └── testutil/
├── migrations/
├── docs/
├── scripts/
├── .github/
│   └── workflows/
├── AGENTS.md
├── CODEX.md
├── ROADMAP.md
├── README.md
├── Makefile
├── Dockerfile
├── compose.yaml
├── go.mod
└── go.sum
```

A estrutura pode evoluir, mas alterações relevantes precisam ser justificadas por uma necessidade real.

## 7. Modelo de domínio

### Session

Representa uma sessão completa de Planning Poker.

Campos conceituais:

- ID;
- Guild ID;
- Channel ID;
- Thread ID;
- Message ID principal;
- URL da tarefa;
- título da tarefa;
- criador;
- facilitador;
- escala;
- estado;
- número da rodada atual;
- estimativa final;
- datas de criação, revelação, conclusão, cancelamento e expiração.

### Round

Representa uma rodada de votação dentro de uma sessão.

Campos conceituais:

- ID;
- Session ID;
- número;
- estado;
- votos;
- estatísticas;
- instante de abertura;
- instante de revelação;
- instante de encerramento.

### Participant

Representa um usuário que deve votar naquela sessão.

Campos conceituais:

- Session ID;
- Discord User ID;
- nome de exibição capturado;
- estado ativo;
- instante de entrada;
- instante de saída.

### Vote

Representa o voto secreto de um participante em uma rodada.

Campos conceituais:

- Round ID;
- Participant ID;
- valor;
- instante do primeiro voto;
- instante da última alteração.

### GuildSettings

Representa configurações específicas de um servidor.

Campos conceituais:

- Guild ID;
- cargo autorizado a facilitar;
- canal padrão;
- escala padrão;
- política de divergência;
- duração até expiração;
- integrações habilitadas.

## 8. Estados

### Sessão

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

### Rodada

```text
OPEN
READY
REVEALED
CLOSED
```

Transições inválidas devem retornar erros de domínio estáveis.

Exemplos:

- não votar antes de entrar;
- não votar após revelação;
- não revelar sem autorização;
- não iniciar nova rodada em sessão concluída;
- não finalizar sem uma estimativa válida;
- não remover participante após a votação, salvo regra explícita.

## 9. Escalas

A escala padrão é Fibonacci modificada:

```text
0, 1, 2, 3, 5, 8, 13, 21, 34, 55, ?, ☕
```

Os valores `?` e `☕` não participam de cálculos numéricos.

Escalas devem ser modeladas como domínio, não como slices espalhados pelo código.

Uma escala deve fornecer:

- valores permitidos;
- posição ordinal;
- validação;
- identificação de valor numérico ou especial;
- formatação para o Discord.

## 10. Divergência

A análise de divergência deve usar posição na escala, não apenas subtração numérica.

Regra inicial:

1. considere somente votos numéricos;
2. calcule a mediana ordinal;
3. calcule a distância de cada voto para a posição mediana;
4. marque como divergente quem estiver a pelo menos duas posições da mediana;
5. quando a amplitude ordinal total for relevante, mencione também os participantes dos extremos;
6. não declare que um voto está errado;
7. solicite contexto, riscos, dependências ou simplificações percebidas.

A política deve ser testável e substituível por configuração futura.

## 11. Regras do Discord

- use comandos de aplicação, preferencialmente slash commands;
- use respostas efêmeras para confirmações privadas de voto;
- nunca publique um valor antes da revelação;
- use componentes interativos com IDs opacos;
- não coloque voto, token ou informação sensível em `custom_id`;
- valide autorização no backend em toda interação;
- não confie apenas na visibilidade de botões;
- trate interações repetidas de forma idempotente;
- considere expiração e indisponibilidade de componentes antigos;
- use threads como contêiner da sessão quando o canal suportar;
- mantenha uma mensagem principal editável com o estado atual;
- o bot deve funcionar após reinício sem perder sessões abertas;
- registre comandos por guild em desenvolvimento e globalmente em produção;
- solicite somente intents e permissões necessárias.

## 12. Persistência e concorrência

PostgreSQL é a fonte de verdade.

Regras:

- toda mutação relevante ocorre em transação;
- voto é único por rodada e participante;
- alteração de voto faz upsert ou update controlado;
- revelação deve bloquear alterações posteriores;
- duas revelações concorrentes não podem duplicar efeitos;
- dois últimos votos concorrentes não podem gerar estado inconsistente;
- use locking otimista por versão ou locking transacional quando necessário;
- nenhuma sessão depende somente da memória do processo;
- mensagens e IDs externos devem ser persistidos;
- migrations devem suportar deploy progressivo.

Restrições recomendadas:

```text
UNIQUE(session_id, discord_user_id)
UNIQUE(round_id, discord_user_id)
UNIQUE(session_id, round_number)
```

## 13. Segurança

- token do Discord nunca entra no Git;
- secrets vêm do ambiente ou de secret manager;
- logs não contêm tokens, payloads secretos nem votos antes da revelação;
- URLs de issues privadas não devem ser enriquecidas sem autorização;
- integrações externas usam menor privilégio;
- todas as ações administrativas validam guild, sessão e ator;
- IDs fornecidos por interação devem ser validados;
- SQL deve ser parametrizado;
- mensagens externas devem ser escapadas ou formatadas com segurança;
- menções devem ser controladas para evitar `@everyone` ou `@here`;
- limite tamanho de texto, título e URL;
- aplique rate limiting interno em ações abusáveis;
- callbacks e goroutines devem respeitar cancelamento.

## 14. Observabilidade

Use `log/slog`.

Campos mínimos quando aplicáveis:

```text
operation
guild_id
channel_id
thread_id
session_id
round_id
interaction_id
actor_discord_id
duration_ms
result
error_code
```

Não registrar valor secreto do voto antes da revelação.

Métricas futuras:

- sessões iniciadas;
- sessões concluídas;
- sessões canceladas;
- duração média;
- rodadas por sessão;
- interações com erro;
- latência de comandos;
- falhas da API do Discord;
- sessões expiradas.

## 15. Padrões Go

- versão de Go definida em `go.mod`;
- código formatado com `gofmt`;
- dependências explícitas;
- interfaces pequenas e definidas pelo consumidor;
- construtores para tipos com invariantes;
- erros com wrapping;
- erros de domínio estáveis;
- `context.Context` em todo I/O;
- timeouts em integrações;
- `slog` para logs;
- graceful shutdown;
- `errgroup` para concorrência estruturada;
- concorrência limitada;
- sem globals mutáveis;
- clock e gerador de IDs injetáveis;
- testes table-driven;
- `go test -race`;
- abstrações somente quando houver caso real.

## 16. Estilo de código

Preferir código com respiro visual entre preparação, execução, validação de erro e retorno.

Preferido:

```go
result, err := service.Execute(ctx, input)

if err != nil {
	return Output{}, fmt.Errorf("execute service: %w", err)
}

return result, nil
```

Evitar:

```go
if result, err := service.Execute(ctx, input); err != nil {
	return Output{}, err
} else {
	return result, nil
}
```

As variáveis podem estar juntas, mas sempre ter quebra de linha entre variável e um bloco if ou qualquer outro bloco.

Outras regras:

- evitar funções longas;
- evitar nomes genéricos como `data`, `manager`, `helper` e `utils`;
- não criar pacote `common` sem responsabilidade clara;
- não compactar lógica apenas para reduzir linhas;
- comentários explicam intenção e decisão, não repetem o código;
- nomes devem refletir a linguagem do domínio;
- pacotes usam nomes curtos e específicos;
- tipos exportados precisam de motivo para serem exportados.

## 17. Erros

Erros de domínio devem ser identificáveis com `errors.Is`.

Exemplos:

```go
var (
	ErrSessionNotFound        = errors.New("planning session not found")
	ErrSessionNotOpen         = errors.New("planning session is not open")
	ErrParticipantNotFound    = errors.New("participant not found")
	ErrAlreadyParticipant     = errors.New("user already participates")
	ErrVoteNotAllowed         = errors.New("vote is not allowed")
	ErrInvalidEstimate        = errors.New("invalid estimate")
	ErrRevealNotAllowed       = errors.New("reveal is not allowed")
	ErrUnauthorizedFacilitator = errors.New("unauthorized facilitator")
)
```

Adaptadores convertem erros do domínio em respostas apropriadas, sem fazer o domínio conhecer Discord.

## 18. Testes

Cada caso de uso deve cobrir:

- caminho feliz;
- estado inválido;
- autorização;
- entrada inválida;
- idempotência;
- concorrência relevante;
- erro de persistência;
- cancelamento de contexto.

Prioridades:

1. testes unitários do domínio;
2. testes unitários dos casos de uso;
3. testes de integração dos repositórios PostgreSQL;
4. testes dos handlers com payloads simulados;
5. poucos testes end-to-end em servidor de desenvolvimento.

Comandos mínimos:

```bash
go test ./...
go test -race ./...
go vet ./...
```

Quando configurado:

```bash
golangci-lint run
```

## 19. Banco e migrations

Migrations são versionadas e nunca reescritas depois de aplicadas em ambiente compartilhado.

Toda migration deve:

- ter `up` e `down`, quando reversão for segura;
- preservar dados existentes;
- evitar locks longos;
- adicionar constraints depois de preparar dados;
- ser compatível com a versão anterior do aplicativo durante rollout;
- incluir índices necessários;
- usar timestamps com timezone;
- armazenar instantes em UTC.

## 20. Commits e pull requests

Commits devem ser pequenos e lógicos.

Exemplos:

```text
feat(planning): add session state transitions
feat(discord): register planning slash command
fix(voting): prevent vote after reveal
test(postgres): cover concurrent vote upsert
docs: document divergence policy
```

Uma pull request deve conter:

- problema;
- solução;
- decisões;
- testes executados;
- riscos;
- migrations;
- screenshots ou exemplos de interação, quando aplicável;
- pendências conhecidas.

## 21. Critério de conclusão

Uma issue só está concluída quando:

- comportamento solicitado foi implementado;
- regras de domínio foram preservadas;
- erros foram tratados;
- testes relevantes passaram;
- race detector foi considerado;
- documentação afetada foi atualizada;
- secrets não foram adicionados;
- logs não expõem votos;
- alterações extras não relacionadas foram evitadas.

# ROADMAP.md — Planning Poker Discord Bot

github: https://github.com/titodelerinofilho/df-planning-poker-discord-bot
github remote: git remote add origin git@github.com:titodelerinofilho/df-planning-poker-discord-bot.git

## Visão do produto

Criar um bot de Planning Poker nativo para Discord que permita a times de software estimarem issues e tarefas sem sair do ambiente de comunicação.

O produto deve priorizar:

- fluxo simples;
- votos realmente secretos;
- recuperação após reinício;
- regras claras;
- boa experiência em threads;
- autorização;
- idempotência;
- histórico;
- integração progressiva com ferramentas de gestão.

## Princípios do roadmap

- cada item implementável deve virar uma issue separada;
- épicos não devem ser enviados inteiros ao agente;
- o MVP deve funcionar sem painel web;
- PostgreSQL é a fonte de verdade desde a primeira sessão persistente;
- regras do domínio precedem melhorias visuais;
- integrações externas entram depois do fluxo principal;
- métricas e operação fazem parte do produto, não são acabamento opcional;
- funcionalidades futuras não devem contaminar a arquitetura atual.

---

# Fase 0 — Fundação do repositório

## Objetivo

Preparar o projeto Go para desenvolvimento seguro, testes e deploy.

### Épico 0.1 — Bootstrap

#### Issue 0.1.1 — Inicializar módulo Go

**Escopo**

- criar `go.mod`;
- definir versão de Go;
- criar `cmd/bot/main.go`;
- adicionar comando de startup mínimo;
- configurar `gofmt`.

**Aceite**

- projeto compila;
- processo inicia e encerra corretamente.

#### Issue 0.1.2 — Configuração tipada

**Escopo**

- criar pacote de configuração;
- carregar variáveis de ambiente;
- validar campos obrigatórios;
- criar `.env.example`;
- impedir impressão de secrets.

**Aceite**

- startup falha com mensagem clara quando configuração obrigatória está ausente.

#### Issue 0.1.3 — Logging estruturado

**Escopo**

- configurar `slog`;
- níveis por ambiente;
- campos de versão e ambiente;
- padronizar logging de erro.

#### Issue 0.1.4 — Graceful shutdown

**Escopo**

- tratar `SIGINT` e `SIGTERM`;
- cancelar contexto raiz;
- encerrar Discord e banco;
- timeout de shutdown.

#### Issue 0.1.5 — Makefile e comandos de qualidade

**Escopo**

- `make run`;
- `make test`;
- `make test-race`;
- `make vet`;
- `make lint`;
- `make build`.

### Épico 0.2 — Qualidade e CI

#### Issue 0.2.1 — Configurar linter

- adicionar `golangci-lint`;
- regras compatíveis com estilo do projeto;
- documentar exceções.

#### Issue 0.2.2 — GitHub Actions

- build;
- testes;
- race detector;
- vet;
- lint;
- cache de módulos.

#### Issue 0.2.3 — Dockerfile

- multi-stage;
- imagem final mínima;
- usuário não root;
- healthcheck quando aplicável.

#### Issue 0.2.4 — Compose local

- bot;
- PostgreSQL;
- volume;
- healthcheck;
- variáveis locais.

---

# Fase 1 — Integração básica com Discord

## Objetivo

Conectar o processo ao Discord e responder a um comando de teste.

### Épico 1.1 — Cliente Discord

#### Issue 1.1.1 — Inicializar `discordgo`

- abrir sessão;
- configurar intents mínimos;
- tratar erro de conexão;
- encerrar sessão no shutdown.

#### Issue 1.1.2 — Registrar comandos por guild

- modo de desenvolvimento;
- Application ID;
- Guild ID;
- criação e remoção controlada de comandos.

#### Issue 1.1.3 — Comando `/ping`

- responder de forma efêmera;
- mostrar latência básica;
- adicionar teste do handler.

#### Issue 1.1.4 — Registro global para produção

- modo configurável;
- documentação de propagação;
- não registrar comandos a cada interação.

### Épico 1.2 — Base de interações

#### Issue 1.2.1 — Router de comandos

- mapear comando e subcomando;
- evitar `switch` gigante;
- erro padronizado.

#### Issue 1.2.2 — Router de componentes

- interpretar IDs opacos;
- validar formato;
- resolver ação e recurso;
- recusar IDs inválidos.

#### Issue 1.2.3 — Presenter Discord

- embeds;
- mensagens efêmeras;
- mensagens públicas;
- mensagens de erro;
- controle de menções.

---

# Fase 2 — Domínio de Planning Poker

## Objetivo

Implementar regras independentes do Discord e do banco.

### Épico 2.1 — Escala

#### Issue 2.1.1 — Escala Fibonacci modificada

Valores:

```text
0, 1, 2, 3, 5, 8, 13, 21, 34, 55, ?, ☕
```

**Aceite**

- valida valor;
- retorna posição ordinal;
- identifica valor especial;
- possui testes table-driven.

#### Issue 2.1.2 — Estatísticas numéricas

- mínimo;
- máximo;
- moda;
- mediana;
- ignorar valores especiais;
- representar ausência de resultado numérico.

### Épico 2.2 — Máquina de estados

#### Issue 2.2.1 — Criar sessão

- identidade;
- criador;
- guild;
- canal;
- thread;
- tarefa;
- escala;
- estado inicial.

#### Issue 2.2.2 — Gerenciar participantes

- entrar;
- sair;
- impedir duplicidade;
- impedir entrada em estado inválido.

#### Issue 2.2.3 — Fechar participantes e abrir votação

- exigir pelo menos quantidade mínima configurada;
- mudar estado;
- criar primeira rodada.

#### Issue 2.2.4 — Registrar e alterar voto

- validar participante;
- validar escala;
- permitir alteração antes da revelação;
- impedir após revelação;
- detectar quando todos votaram.

#### Issue 2.2.5 — Revelar rodada

- validar estado;
- congelar votos;
- calcular estatísticas;
- produzir resultado estruturado.

#### Issue 2.2.6 — Nova rodada

- fechar rodada anterior;
- criar rodada seguinte;
- manter participantes;
- limpar votos ativos.

#### Issue 2.2.7 — Finalizar sessão

- validar estimativa final;
- registrar conclusão;
- impedir mutações futuras.

#### Issue 2.2.8 — Cancelar e expirar

- motivo;
- ator;
- instante;
- estados permitidos.

### Épico 2.3 — Divergência

#### Issue 2.3.1 — Distância ordinal

- calcular mediana por posição;
- calcular distância dos votos;
- ignorar especiais.

#### Issue 2.3.2 — Detectar extremos e discrepantes

- limiar padrão de duas posições;
- participantes de mínimo e máximo;
- resultado estruturado.

#### Issue 2.3.3 — Política configurável

- limiar por guild;
- ativar ou desativar menção;
- preservar regra padrão.

---

# Fase 3 — Persistência PostgreSQL

## Objetivo

Garantir recuperação, consistência e concorrência.

### Épico 3.1 — Banco e migrations

#### Issue 3.1.1 — Conexão com PostgreSQL

- `pgxpool`;
- timeout;
- health check;
- shutdown.

#### Issue 3.1.2 — Ferramenta de migrations

- escolher uma;
- criar comando;
- documentar fluxo;
- CI valida migrations.

#### Issue 3.1.3 — Migration de sessões

Campos mínimos:

```text
id
guild_id
channel_id
thread_id
message_id
issue_url
issue_title
created_by_discord_id
facilitator_discord_id
status
scale
current_round_number
final_estimate
version
created_at
updated_at
revealed_at
completed_at
cancelled_at
expires_at
```

#### Issue 3.1.4 — Migration de rodadas

```text
id
session_id
round_number
status
opened_at
ready_at
revealed_at
closed_at
statistics
```

#### Issue 3.1.5 — Migration de participantes

```text
id
session_id
discord_user_id
display_name
active
joined_at
left_at
```

#### Issue 3.1.6 — Migration de votos

```text
id
round_id
discord_user_id
estimate
created_at
updated_at
```

#### Issue 3.1.7 — Migration de configurações por guild

- cargo facilitador;
- canal padrão;
- escala;
- limiar;
- expiração.

### Épico 3.2 — Repositórios

#### Issue 3.2.1 — Repositório de sessão

- criar;
- buscar por ID;
- buscar por thread;
- atualizar com versão;
- listar abertas.

#### Issue 3.2.2 — Repositório de participante

- adicionar;
- remover;
- listar;
- validar duplicidade.

#### Issue 3.2.3 — Repositório de rodada e voto

- criar rodada;
- upsert de voto;
- buscar votos;
- congelar rodada.

#### Issue 3.2.4 — Transação da aplicação

- porta de transação;
- implementação PostgreSQL;
- casos de uso atômicos.

### Épico 3.3 — Concorrência e idempotência

#### Issue 3.3.1 — Dois últimos votos concorrentes

- teste de integração;
- estado final consistente;
- uma única transição para `READY`.

#### Issue 3.3.2 — Revelação concorrente

- duas solicitações;
- somente um efeito lógico;
- segunda resposta idempotente.

#### Issue 3.3.3 — Voto concorrente com revelação

- bloquear alteração após congelamento;
- definir ordem transacional;
- testar race no banco.

#### Issue 3.3.4 — Chave de idempotência de interação

- persistir interações processadas quando necessário;
- tratar retries;
- evitar efeitos duplicados.

---

# Fase 4 — MVP funcional no Discord

## Objetivo

Entregar o fluxo completo de uma sessão dentro de uma thread.

### Épico 4.1 — Iniciar sessão

#### Issue 4.1.1 — `/planning start`

Opções:

```text
issue
title opcional
scale opcional
```

**Aceite**

- valida URL;
- cria sessão;
- cria thread;
- persiste IDs;
- publica mensagem principal.

#### Issue 4.1.2 — Resolver título básico

- usar título informado;
- caso ausente, exibir domínio e identificador da URL;
- não consultar provedor ainda.

#### Issue 4.1.3 — Criar thread

- validar canal;
- nome seguro;
- fallback documentado quando thread não for suportada;
- persistir Thread ID.

### Épico 4.2 — Participantes

#### Issue 4.2.1 — Botão Participar

- adiciona usuário;
- resposta efêmera;
- atualiza mensagem principal.

#### Issue 4.2.2 — Botão Sair

- permitido durante `JOINING`;
- atualiza contagem;
- resposta efêmera.

#### Issue 4.2.3 — Fechar participantes

- somente facilitador;
- abre votação;
- desabilita entrada e saída.

#### Issue 4.2.4 — Adição manual

- opção de mencionar usuários;
- validar bots;
- evitar duplicidade.

### Épico 4.3 — Votação

#### Issue 4.3.1 — Abrir seletor de voto

- select menu ou modal adequado;
- valores da escala;
- identificar sessão de forma opaca.

#### Issue 4.3.2 — Registrar voto secreto

- persistir;
- responder efemeramente;
- não logar valor;
- permitir alteração.

#### Issue 4.3.3 — Atualizar status público

Exibir:

- participantes;
- quem votou;
- quem está pendente;
- contagem;
- nunca valores.

#### Issue 4.3.4 — Todos votaram

- mudar estado para `READY_TO_REVEAL`;
- editar mensagem;
- mencionar facilitador;
- evitar aviso duplicado.

#### Issue 4.3.5 — `/planning status`

- localizar sessão pela thread;
- exibir estado;
- participantes;
- pendentes;
- rodada atual.

### Épico 4.4 — Revelação

#### Issue 4.4.1 — `/planning reveal`

- somente facilitador ou cargo autorizado;
- validar todos votaram ou política configurada;
- revelar de forma idempotente.

#### Issue 4.4.2 — Exibir resultado

- usuário e voto;
- mínimo;
- máximo;
- moda;
- mediana;
- valores especiais;
- rodada.

#### Issue 4.4.3 — Sinalizar divergência

- mencionar extremos ou discrepantes;
- pedir explicação sobre riscos, dependências ou simplificações;
- controlar allowed mentions.

### Épico 4.5 — Revotação e conclusão

#### Issue 4.5.1 — `/planning revote`

- somente facilitador;
- iniciar nova rodada;
- manter participantes;
- atualizar mensagem.

#### Issue 4.5.2 — Histórico resumido de rodadas

- listar estatísticas por rodada;
- não criar mensagem excessivamente longa.

#### Issue 4.5.3 — `/planning finish`

Opção:

```text
estimate
```

- validar escala ou política;
- concluir;
- desabilitar componentes;
- publicar resultado final.

#### Issue 4.5.4 — `/planning cancel`

- motivo opcional;
- concluir como cancelada;
- desabilitar componentes.

---

# Marco MVP

O MVP é considerado entregue quando:

- bot pode ser instalado em servidor de teste;
- `/planning start` cria sessão e thread;
- usuários entram;
- facilitador fecha participantes;
- participantes votam secretamente;
- status público não revela valores;
- bot avisa quando todos votaram;
- facilitador revela;
- divergência é apresentada;
- nova rodada funciona;
- sessão pode ser finalizada;
- reinício do bot não perde estado;
- concorrência crítica possui testes;
- deploy documentado.

---

# Fase 5 — Configuração por servidor

## Objetivo

Permitir que cada guild adapte o comportamento.

### Épico 5.1 — Permissões

#### Issue 5.1.1 — Cargo de facilitador

- configurar cargo;
- criador da sessão;
- administradores;
- precedência documentada.

#### Issue 5.1.2 — Política de revelação

Opções futuras:

```text
only_when_all_voted
facilitator_can_force
majority_threshold
```

#### Issue 5.1.3 — Canal permitido

- canal padrão;
- allowlist opcional;
- mensagem de erro útil.

### Épico 5.2 — Preferências

#### Issue 5.2.1 — Escala padrão

- Fibonacci;
- T-shirt sizes futura;
- escala customizada futura.

#### Issue 5.2.2 — Limiar de divergência

- distância ordinal;
- menção de extremos;
- desligar aviso.

#### Issue 5.2.3 — Expiração

- duração padrão;
- lembretes;
- encerramento automático.

#### Issue 5.2.4 — Localização

- português;
- inglês;
- catálogo de mensagens;
- fallback.

### Épico 5.3 — `/planning configure`

- visualizar configuração;
- alterar opções autorizadas;
- validação de permissão;
- auditoria.

---

# Fase 6 — Integração com provedores de issues

## Objetivo

Enriquecer a sessão sem tornar a integração requisito para votar.

### Épico 6.1 — Contrato de provedor

#### Issue 6.1.1 — Porta `IssueProvider`

Saída estruturada:

```text
provider
external_id
title
description_summary
status
labels
assignees
project
url
```

#### Issue 6.1.2 — Resolver provedor por URL

- GitHub;
- GitLab;
- Jira;
- fallback genérico.

#### Issue 6.1.3 — Tratamento de issue privada

- credencial por guild ou instalação;
- mensagem segura;
- não bloquear sessão quando enriquecimento falhar.

### Épico 6.2 — GitHub

#### Issue 6.2.1 — Issues públicas

- título;
- estado;
- labels;
- assignees.

#### Issue 6.2.2 — GitHub App

- instalação;
- armazenamento seguro;
- acesso mínimo;
- repositórios privados.

#### Issue 6.2.3 — Atualizar estimativa na issue

- recurso opt-in;
- label, comentário ou campo de projeto;
- confirmação explícita;
- auditoria;
- idempotência.

### Épico 6.3 — GitLab

- issue pública;
- token por grupo;
- projetos privados;
- comentário ou weight.

### Épico 6.4 — Jira

- conexão;
- leitura de issue;
- atualização de story points;
- mapeamento de campos;
- auditoria.

---

# Fase 7 — Operação, reconciliação e confiabilidade

## Objetivo

Manter mensagens e banco consistentes diante de falhas.

### Épico 7.1 — Caixa de efeitos

#### Issue 7.1.1 — Persistir efeitos pendentes

Tipos:

```text
CREATE_THREAD
CREATE_MAIN_MESSAGE
UPDATE_MAIN_MESSAGE
SEND_READY_NOTIFICATION
SEND_REVEAL_RESULT
DISABLE_COMPONENTS
```

#### Issue 7.1.2 — Worker interno

- polling limitado;
- retry com backoff;
- idempotência;
- dead-letter lógica;
- shutdown.

#### Issue 7.1.3 — Reconciliação

- reconstruir mensagem a partir do banco;
- detectar mensagem removida;
- recriar quando permitido.

### Épico 7.2 — Expiração e lembretes

#### Issue 7.2.1 — Worker de expiração

- localizar sessões vencidas;
- expirar;
- desabilitar componentes.

#### Issue 7.2.2 — Lembrete de pendentes

- intervalo configurável;
- evitar spam;
- menções controladas.

### Épico 7.3 — Rate limiting e resiliência

- retry respeitando limites do Discord;
- circuit breaker somente se justificado;
- timeouts;
- métricas de falha.

---

# Fase 8 — Auditoria e histórico

## Objetivo

Permitir rastrear decisões sem expor indevidamente votos secretos.

### Épico 8.1 — Eventos de auditoria

Eventos:

```text
SESSION_STARTED
PARTICIPANT_JOINED
PARTICIPANT_LEFT
PARTICIPANTS_CLOSED
VOTE_CAST
VOTE_CHANGED
ROUND_READY
ROUND_REVEALED
ROUND_RESTARTED
SESSION_COMPLETED
SESSION_CANCELLED
SESSION_EXPIRED
SETTINGS_CHANGED
EXTERNAL_ISSUE_UPDATED
```

Antes da revelação, evento de voto não registra valor em payload de auditoria acessível.

### Épico 8.2 — Comandos de histórico

#### Issue 8.2.1 — `/planning history`

- últimas sessões;
- filtros básicos;
- paginação.

#### Issue 8.2.2 — `/planning show`

- sessão específica;
- rodadas;
- resultado final;
- link da thread.

#### Issue 8.2.3 — Exportação

- JSON ou CSV;
- permissão;
- limites;
- proteção de dados.

---

# Fase 9 — Experiência avançada

## Objetivo

Aprimorar usabilidade após estabilidade do núcleo.

### Épico 9.1 — Templates de sessão

- escala;
- participantes por cargo;
- política de divergência;
- canal.

### Épico 9.2 — Participantes por cargo

- selecionar cargo;
- excluir bots;
- snapshot no início;
- tratar membros indisponíveis.

### Épico 9.3 — Sessão ad hoc

- sem URL;
- título obrigatório;
- descrição curta.

### Épico 9.4 — Timer

- duração da votação;
- contagem;
- lembretes;
- ação ao expirar.

### Épico 9.5 — Observador

- usuário acompanha;
- não entra na contagem;
- não vota.

### Épico 9.6 — Consenso

- política de consenso;
- unanimidade;
- amplitude máxima;
- sugestão sem decisão automática.

---

# Fase 10 — Painel administrativo opcional

## Objetivo

Adicionar painel somente depois da maturidade do bot.

### Escopo futuro

- login via Discord OAuth2;
- guilds administradas;
- configurações;
- histórico;
- integrações;
- métricas;
- auditoria.

### Restrições

- painel não acessa tabelas sem camada de aplicação;
- não duplicar regra de negócio;
- autorização por guild;
- CSRF e sessão segura;
- secrets criptografados.

---

# Fase 11 — Distribuição pública

## Objetivo

Preparar o bot para instalação por múltiplos servidores.

### Épico 11.1 — Instalação

- documentação;
- escopos mínimos;
- permissões mínimas;
- política de privacidade;
- termos;
- suporte.

### Épico 11.2 — Escalabilidade

- sharding somente quando necessário;
- múltiplas réplicas;
- liderança de workers;
- locks distribuídos via PostgreSQL inicialmente;
- métricas por guild.

### Épico 11.3 — Limites e planos futuros

Somente caso vire produto comercial:

- limites por guild;
- recursos gratuitos;
- recursos pagos;
- billing;
- entitlement;
- sem prejudicar segurança do núcleo.

---

# Backlog técnico transversal

## Segurança

- secret scanning;
- dependabot;
- atualização de dependências;
- proteção contra menções abusivas;
- validação de URL;
- menor privilégio;
- threat model.

## Observabilidade

- métricas Prometheus ou OpenTelemetry;
- tracing de casos de uso;
- dashboards;
- alertas;
- correlação por interação.

## Performance

- índices por consultas reais;
- paginação;
- limitar payload;
- benchmark apenas em áreas críticas;
- evitar otimização prematura.

## Documentação

- arquitetura;
- domínio;
- estados;
- comandos;
- setup Discord;
- banco;
- deploy;
- integração;
- segurança;
- troubleshooting.

## Testes

- unitários;
- integração;
- concorrência;
- contrato de provedores;
- smoke test;
- end-to-end controlado.

---

# Ordem sugerida de implementação

```text
Fase 0
  ↓
Fase 1
  ↓
Fase 2
  ↓
Fase 3
  ↓
Fase 4
  ↓
MARCO MVP
  ↓
Fase 5
  ↓
Fase 6
  ↓
Fases 7 e 8
  ↓
Fase 9
  ↓
Fases 10 e 11, somente se houver necessidade comercial
```

## Primeiras 15 issues recomendadas

1. Inicializar módulo Go.
2. Criar configuração tipada.
3. Configurar `slog`.
4. Implementar graceful shutdown.
5. Configurar CI.
6. Integrar `discordgo`.
7. Registrar `/ping` por guild.
8. Implementar escala Fibonacci.
9. Implementar entidade e estados de sessão.
10. Implementar participantes.
11. Implementar rodada e voto secreto no domínio.
12. Implementar estatísticas e divergência.
13. Configurar PostgreSQL e migrations.
14. Implementar repositório de sessão.
15. Implementar `/planning start`.

Essas issues devem ser abertas separadamente e executadas uma por vez.

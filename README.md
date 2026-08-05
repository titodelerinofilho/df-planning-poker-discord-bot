# DF Planning Poker Discord Bot

Bot de Planning Poker para Discord escrito em Go.

## Desenvolvimento

Este repositório segue o fluxo descrito em `AGENTS.md`, `CODEX.md` e `ROADMAP.md`: uma issue por vez, monólito modular e regras de domínio isoladas de Discord e PostgreSQL.

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

Os logs são emitidos em JSON via `slog` e incluem os campos fixos `version`, `environment`, `channel` e `correlation_identifier`.

Os canais seguem esta separação:

- `exceptions`: STDERR, nível `CRITICAL`.
- `requests`: STDOUT, nível `INFO`.
- `responses`: STDOUT, nível `INFO`.
- `database_queries`: STDOUT, nível `INFO`.

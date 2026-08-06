package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	discordadapter "github.com/titodelerinofilho/df-planning-poker-discord-bot/internal/adapters/discord"
)

func TestRunLogsStartupMessage(t *testing.T) {
	setRequiredEnv(t)

	stdoutFile := newLogFile(t)
	stderr := newLogFile(t)
	ctx, cancel := context.WithCancel(context.Background())
	stdout := &cancelingWriter{
		writer: stdoutFile,
		cancel: cancel,
		after:  6,
	}
	dependencies := runtimeDependencies{
		newDiscordBot: func(string) (discordBot, error) {
			return &fakeDiscordBot{}, nil
		},
	}

	err := run(ctx, stdout, stderr, dependencies)

	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	stderrOutput := readLogFile(t, stderr)

	if stderrOutput != "" {
		t.Fatalf("stderr = %q, want no output", stderrOutput)
	}

	stdoutOutput := readLogFile(t, stdoutFile)
	entries := decodeMainLogEntries(t, stdoutOutput)

	if len(entries) != 9 {
		t.Fatalf("log entries = %d, want 9", len(entries))
	}

	if entries[0]["channel"] != "requests" {
		t.Fatalf("first channel = %v, want requests", entries[0]["channel"])
	}

	if entries[0]["level"] != "INFO" {
		t.Fatalf("first level = %v, want INFO", entries[0]["level"])
	}

	if entries[0]["correlation_identifier"] != startupCorrelationID {
		t.Fatalf("first correlation_identifier = %v, want %s", entries[0]["correlation_identifier"], startupCorrelationID)
	}

	if entries[1]["channel"] != "responses" {
		t.Fatalf("second channel = %v, want responses", entries[1]["channel"])
	}

	if entries[1]["level"] != "INFO" {
		t.Fatalf("second level = %v, want INFO", entries[1]["level"])
	}

	if entries[1]["environment"] != "development" {
		t.Fatalf("environment = %v, want development", entries[1]["environment"])
	}

	if entries[1]["version"] != appVersion {
		t.Fatalf("version = %v, want %s", entries[1]["version"], appVersion)
	}

	if entries[1]["correlation_identifier"] != startupCorrelationID {
		t.Fatalf("second correlation_identifier = %v, want %s", entries[1]["correlation_identifier"], startupCorrelationID)
	}

	content, ok := entries[1]["content"].(map[string]any)

	if !ok {
		t.Fatalf("content = %T, want object", entries[1]["content"])
	}

	if content["message"] != startupMessage {
		t.Fatalf("content.message = %v, want %s", content["message"], startupMessage)
	}

	if content["command_registration_mode"] != "guild" {
		t.Fatalf("content.command_registration_mode = %v, want guild", content["command_registration_mode"])
	}

	if entries[2]["channel"] != "requests" {
		t.Fatalf("third channel = %v, want requests", entries[2]["channel"])
	}

	if entries[2]["correlation_identifier"] != startupCorrelationID {
		t.Fatalf("third correlation_identifier = %v, want %s", entries[2]["correlation_identifier"], startupCorrelationID)
	}

	if entries[3]["channel"] != "responses" {
		t.Fatalf("fourth channel = %v, want responses", entries[3]["channel"])
	}

	if entries[3]["correlation_identifier"] != startupCorrelationID {
		t.Fatalf("fourth correlation_identifier = %v, want %s", entries[3]["correlation_identifier"], startupCorrelationID)
	}

	if entries[4]["channel"] != "requests" {
		t.Fatalf("fifth channel = %v, want requests", entries[4]["channel"])
	}

	if entries[4]["correlation_identifier"] != startupCorrelationID {
		t.Fatalf("fifth correlation_identifier = %v, want %s", entries[4]["correlation_identifier"], startupCorrelationID)
	}

	if entries[5]["channel"] != "responses" {
		t.Fatalf("sixth channel = %v, want responses", entries[5]["channel"])
	}

	if entries[5]["correlation_identifier"] != startupCorrelationID {
		t.Fatalf("sixth correlation_identifier = %v, want %s", entries[5]["correlation_identifier"], startupCorrelationID)
	}

	if entries[6]["channel"] != "requests" {
		t.Fatalf("seventh channel = %v, want requests", entries[6]["channel"])
	}

	if entries[6]["correlation_identifier"] != shutdownCorrelationID {
		t.Fatalf("seventh correlation_identifier = %v, want %s", entries[6]["correlation_identifier"], shutdownCorrelationID)
	}

	if entries[7]["channel"] != "responses" {
		t.Fatalf("eighth channel = %v, want responses", entries[7]["channel"])
	}

	if entries[7]["correlation_identifier"] != shutdownCorrelationID {
		t.Fatalf("eighth correlation_identifier = %v, want %s", entries[7]["correlation_identifier"], shutdownCorrelationID)
	}

	if entries[8]["channel"] != "responses" {
		t.Fatalf("ninth channel = %v, want responses", entries[8]["channel"])
	}

	if entries[8]["correlation_identifier"] != shutdownCorrelationID {
		t.Fatalf("ninth correlation_identifier = %v, want %s", entries[8]["correlation_identifier"], shutdownCorrelationID)
	}
}

func TestRunReturnsConfigurationError(t *testing.T) {
	unsetRequiredEnv(t)

	stdout := newLogFile(t)
	stderr := newLogFile(t)

	err := run(context.Background(), stdout, stderr, testRuntimeDependencies())

	if err == nil {
		t.Fatal("run() error = nil, want configuration error")
	}

	stdoutOutput := readLogFile(t, stdout)

	if stdoutOutput != "" {
		t.Fatalf("run() wrote stdout %q, want no output", stdoutOutput)
	}

	stderrOutput := readLogFile(t, stderr)

	if stderrOutput != "" {
		t.Fatalf("run() wrote stderr %q, want no output", stderrOutput)
	}
}

func TestRunReturnsContextError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	stdout := newLogFile(t)
	stderr := newLogFile(t)

	err := run(ctx, stdout, stderr, testRuntimeDependencies())

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("run() error = %v, want context.Canceled", err)
	}

	stdoutOutput := readLogFile(t, stdout)

	if stdoutOutput != "" {
		t.Fatalf("run() wrote stdout %q, want no output", stdoutOutput)
	}

	stderrOutput := readLogFile(t, stderr)

	if stderrOutput != "" {
		t.Fatalf("run() wrote stderr %q, want no output", stderrOutput)
	}
}

func TestRunRejectsInvalidLogLevel(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("LOG_LEVEL", "verbose")

	stdout := newLogFile(t)
	stderr := newLogFile(t)

	err := run(context.Background(), stdout, stderr, testRuntimeDependencies())

	if err == nil {
		t.Fatal("run() error = nil, want invalid log level error")
	}

	if !strings.Contains(err.Error(), "configure logging") {
		t.Fatalf("run() error = %v, want configure logging context", err)
	}
}

func TestRunReturnsDiscordOpenError(t *testing.T) {
	setRequiredEnv(t)

	stdout := newLogFile(t)
	stderr := newLogFile(t)
	openErr := errors.New("open discord")
	dependencies := runtimeDependencies{
		newDiscordBot: func(string) (discordBot, error) {
			return &fakeDiscordBot{openErr: openErr}, nil
		},
	}

	err := run(context.Background(), stdout, stderr, dependencies)

	if !errors.Is(err, openErr) {
		t.Fatalf("run() error = %v, want open discord error", err)
	}
}

func TestRunSyncsGuildCommands(t *testing.T) {
	setRequiredEnv(t)

	stdoutFile := newLogFile(t)
	stderr := newLogFile(t)
	ctx, cancel := context.WithCancel(context.Background())
	stdout := &cancelingWriter{
		writer: stdoutFile,
		cancel: cancel,
		after:  6,
	}
	bot := &fakeDiscordBot{}
	dependencies := runtimeDependencies{
		newDiscordBot: func(string) (discordBot, error) {
			return bot, nil
		},
	}

	err := run(ctx, stdout, stderr, dependencies)

	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	if bot.syncGuildCommandCalls != 1 {
		t.Fatalf("sync guild command calls = %d, want 1", bot.syncGuildCommandCalls)
	}

	if bot.applicationID != "discord-application-id" {
		t.Fatalf("application id = %q, want discord-application-id", bot.applicationID)
	}

	if bot.guildID != "discord-guild-id" {
		t.Fatalf("guild id = %q, want discord-guild-id", bot.guildID)
	}
}

func TestRunSyncsGlobalCommands(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("COMMAND_REGISTRATION_MODE", "global")
	t.Setenv("DISCORD_GUILD_ID", "")

	stdoutFile := newLogFile(t)
	stderr := newLogFile(t)
	ctx, cancel := context.WithCancel(context.Background())
	stdout := &cancelingWriter{
		writer: stdoutFile,
		cancel: cancel,
		after:  6,
	}
	bot := &fakeDiscordBot{}
	dependencies := runtimeDependencies{
		newDiscordBot: func(string) (discordBot, error) {
			return bot, nil
		},
	}

	err := run(ctx, stdout, stderr, dependencies)

	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	if bot.syncGlobalCommandCalls != 1 {
		t.Fatalf("sync global command calls = %d, want 1", bot.syncGlobalCommandCalls)
	}

	if bot.syncGuildCommandCalls != 0 {
		t.Fatalf("sync guild command calls = %d, want 0", bot.syncGuildCommandCalls)
	}

	if bot.applicationID != "discord-application-id" {
		t.Fatalf("application id = %q, want discord-application-id", bot.applicationID)
	}
}

func TestRunReturnsGuildCommandSyncError(t *testing.T) {
	setRequiredEnv(t)

	stdout := newLogFile(t)
	stderr := newLogFile(t)
	syncErr := errors.New("discord api")
	bot := &fakeDiscordBot{syncGuildCommandsErr: syncErr}
	dependencies := runtimeDependencies{
		newDiscordBot: func(string) (discordBot, error) {
			return bot, nil
		},
	}

	err := run(context.Background(), stdout, stderr, dependencies)

	if !errors.Is(err, syncErr) {
		t.Fatalf("run() error = %v, want sync error", err)
	}

	if bot.closeCalls != 1 {
		t.Fatalf("discord close calls = %d, want 1", bot.closeCalls)
	}
}

func TestRunReturnsGlobalCommandSyncError(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("COMMAND_REGISTRATION_MODE", "global")
	t.Setenv("DISCORD_GUILD_ID", "")

	stdout := newLogFile(t)
	stderr := newLogFile(t)
	syncErr := errors.New("discord api")
	bot := &fakeDiscordBot{syncGlobalCommandsErr: syncErr}
	dependencies := runtimeDependencies{
		newDiscordBot: func(string) (discordBot, error) {
			return bot, nil
		},
	}

	err := run(context.Background(), stdout, stderr, dependencies)

	if !errors.Is(err, syncErr) {
		t.Fatalf("run() error = %v, want sync error", err)
	}

	if bot.closeCalls != 1 {
		t.Fatalf("discord close calls = %d, want 1", bot.closeCalls)
	}
}

func TestRunClosesDiscordOnShutdown(t *testing.T) {
	setRequiredEnv(t)

	stdoutFile := newLogFile(t)
	stderr := newLogFile(t)
	ctx, cancel := context.WithCancel(context.Background())
	stdout := &cancelingWriter{
		writer: stdoutFile,
		cancel: cancel,
		after:  6,
	}
	bot := &fakeDiscordBot{}
	dependencies := runtimeDependencies{
		newDiscordBot: func(string) (discordBot, error) {
			return bot, nil
		},
	}

	err := run(ctx, stdout, stderr, dependencies)

	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	if bot.openCalls != 1 {
		t.Fatalf("discord open calls = %d, want 1", bot.openCalls)
	}

	if bot.closeCalls != 1 {
		t.Fatalf("discord close calls = %d, want 1", bot.closeCalls)
	}
}

func testRuntimeDependencies() runtimeDependencies {
	return runtimeDependencies{
		newDiscordBot: func(string) (discordBot, error) {
			return &fakeDiscordBot{}, nil
		},
	}
}

type fakeDiscordBot struct {
	openCalls              int
	closeCalls             int
	syncGuildCommandCalls  int
	syncGlobalCommandCalls int

	applicationID string
	guildID       string

	openErr               error
	closeErr              error
	syncGuildCommandsErr  error
	syncGlobalCommandsErr error
}

func (bot *fakeDiscordBot) Open(context.Context) error {
	bot.openCalls++

	return bot.openErr
}

func (bot *fakeDiscordBot) Close(context.Context) error {
	bot.closeCalls++

	return bot.closeErr
}

func (bot *fakeDiscordBot) SyncGuildCommands(_ context.Context, applicationID string, guildID string, _ []discordadapter.CommandDefinition) error {
	bot.syncGuildCommandCalls++
	bot.applicationID = applicationID
	bot.guildID = guildID

	return bot.syncGuildCommandsErr
}

func (bot *fakeDiscordBot) SyncGlobalCommands(_ context.Context, applicationID string, _ []discordadapter.CommandDefinition) error {
	bot.syncGlobalCommandCalls++
	bot.applicationID = applicationID

	return bot.syncGlobalCommandsErr
}

func setRequiredEnv(t *testing.T) {
	t.Helper()

	t.Setenv("APP_ENV", "development")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("DISCORD_TOKEN", "discord-token")
	t.Setenv("DISCORD_APPLICATION_ID", "discord-application-id")
	t.Setenv("DISCORD_GUILD_ID", "discord-guild-id")
	t.Setenv("DATABASE_URL", "postgres://user:password@localhost:5432/df_poker")
	t.Setenv("COMMAND_REGISTRATION_MODE", "guild")
	t.Setenv("SESSION_EXPIRATION", "24h")
	t.Setenv("SHUTDOWN_TIMEOUT", "10s")
}

func unsetRequiredEnv(t *testing.T) {
	t.Helper()

	names := []string{
		"APP_ENV",
		"LOG_LEVEL",
		"DISCORD_TOKEN",
		"DISCORD_APPLICATION_ID",
		"DISCORD_GUILD_ID",
		"DATABASE_URL",
		"COMMAND_REGISTRATION_MODE",
		"SESSION_EXPIRATION",
		"SHUTDOWN_TIMEOUT",
	}

	original := make(map[string]string, len(names))
	present := make(map[string]bool, len(names))

	for _, name := range names {
		value, ok := os.LookupEnv(name)
		original[name] = value
		present[name] = ok

		err := os.Unsetenv(name)

		if err != nil {
			t.Fatalf("unset %s: %v", name, err)
		}
	}

	t.Cleanup(func() {
		for _, name := range names {
			if !present[name] {
				err := os.Unsetenv(name)

				if err != nil {
					t.Fatalf("restore unset %s: %v", name, err)
				}

				continue
			}

			err := os.Setenv(name, original[name])

			if err != nil {
				t.Fatalf("restore set %s: %v", name, err)
			}
		}
	})
}

func newLogFile(t *testing.T) *os.File {
	t.Helper()

	file, err := os.CreateTemp(t.TempDir(), "bot-log-*.json")

	if err != nil {
		t.Fatalf("create temp log file: %v", err)
	}

	return file
}

type cancelingWriter struct {
	writer io.Writer
	cancel context.CancelFunc
	after  int
	writes int
}

func (writer *cancelingWriter) Write(content []byte) (int, error) {
	written, err := writer.writer.Write(content)

	if err != nil {
		return written, err
	}

	writer.writes++

	if writer.writes == writer.after {
		writer.cancel()
	}

	return written, nil
}

func readLogFile(t *testing.T, file *os.File) string {
	t.Helper()

	_, err := file.Seek(0, io.SeekStart)

	if err != nil {
		t.Fatalf("seek log file: %v", err)
	}

	content, err := io.ReadAll(file)

	if err != nil {
		t.Fatalf("read log file: %v", err)
	}

	return string(content)
}

func decodeMainLogEntries(t *testing.T, raw string) []map[string]any {
	t.Helper()

	lines := strings.Split(strings.TrimSpace(raw), "\n")
	entries := make([]map[string]any, 0, len(lines))

	for _, line := range lines {
		var entry map[string]any

		err := json.Unmarshal([]byte(line), &entry)

		if err != nil {
			t.Fatalf("decode log entry %q: %v", line, err)
		}

		entries = append(entries, entry)
	}

	return entries
}

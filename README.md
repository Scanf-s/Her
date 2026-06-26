# Her

**Speak with AI in English which is fully local, free, and private.**

Her is a desktop class web app for practicing spoken English with an AI partner
that runs **entirely on your own machine**.

No cloud, no Model usage API keys, no cost. 
You just talk. It listens and replies in a natural, flowing conversation. 
The name comes from the film *Her* and its AI companion, `Samantha`.

> **!!Early development stage!!** Text chat and push-to-talk voice **input** work.
> Voice output (TTS), hands-free conversation, and the session feedback report are still going on [Development status](#development-status).

## What it does

- **Text chat**: type in English, replies stream back token-by-token.
- **Voice input**: hold to talk and your speech goes straight to an audio-native model (Gemma 4), no separate speech-to-text step.
- **Samantha persona**: a warm conversation partner that reacts and asks follow-ups, tuned to a CEFR English level.
- **100% local**: your audio never leaves your machine.

## How it works

A small **Go orchestrator** serves a thin web page and a WebSocket endpoint.
Each connection owns one conversation `Session` that keeps history and streams every turn through an `Understander` interface, backed by a local OpenAI-compatible LLM.

```
Browser (mic / text) ---WebSocket-->  Go server ---HTTP /v1/chat/completions-->  Gemma 4
token stream  <------------------  (session + understander)  <---------------  (LiteRT-LM)
```

- `cmd/her`: entrypoint: load config, wire dependencies, serve.
- `internal/config`: configuration and defaults.
- `internal/server`: HTTP + WebSocket handler.
- `internal/session`: per-connection conversation state and token streaming.
- `internal/understander`: the LLM interface and its LiteRT-LM client.
- `web/`: vanilla-JS frontend (push-to-talk + text).

## Getting started

### Prerequisites

- **Go 1.26+**
- A **local OpenAI-compatible LLM server** running an audio-capable **Gemma 4** model on `http://localhost:8081`. 
- This project uses Google's [LiteRT-LM](https://github.com/google-ai-edge/LiteRT-LM) (`litert-lm serve`).
- *(optional)* [goreman](https://github.com/mattn/goreman) for the `make dev`
  shortcut: `go install github.com/mattn/goreman@latest`

### Download the Gemma 4 model

The Gemma models are **gated** on Hugging Face, so you need a (free) account and an
access token.

```bash
# 1. Install the LiteRT-LM CLI (pulls in the `hf` Hugging Face CLI as a dependency)
pip install litert-lm

# 2. Accept the license once, signed in, at the model page:
#    https://huggingface.co/litert-community/gemma-4-E4B-it-litert-lm

# 3. Log in with a "read" token from https://huggingface.co/settings/tokens
hf auth login                       # paste the token when prompted
# (or non-interactively: export HF_TOKEN=hf_xxxxxxxxxxxxxxxxxxxx)

# 4. Download + register the model under the name Her expects (gemma4-e4b)
litert-lm import \
  --from-huggingface-repo=litert-community/gemma-4-E4B-it-litert-lm \
  model.litertlm \
  gemma4-e4b
```

`litert-lm serve --port 8081` then serves the imported model, and Her reaches it via the default `HER_LLM_MODEL=gemma4-e4b`.
Lighter/heavier variants exist too -> `gemma-4-E2B-it-litert-lm` (smaller) and `gemma-4-12B-it-litert-lm` (desktop, more
capable).
You can swap the model by modifying `HER_LLM_MODEL` configuration.

> **Voice input** needs the CPU audio backend for this model. 
> If a voice turn fails with a backend error, set `HER_LLM_MODEL=gemma4-e4b,cpu`.

### Run

**Option A: Everything at once** (needs `goreman` and `litert-lm` on your device):

```sh
make dev   # starts the LLM server + web app together (see Procfile)
```

**Option B: manual**

```sh
# 1. start your Gemma 4 server on http://localhost:8081, then:
make build && ./her      # or: go run ./cmd/her
```

Open <http://localhost:8080>, then **hold the green button to talk** or type a message and press Send.

### Test

```sh
make test   # go test ./...
```

## Configuration

All settings are read from environment variables with sensible defaults:

| Variable                 | Default                  | Description                              |
| ------------------------ | ------------------------ | ---------------------------------------- |
| `HER_ADDR`               | `:8080`                  | Web server listen address                |
| `HER_LLM_ENDPOINT`       | `http://localhost:8081`  | OpenAI-compatible LLM base URL           |
| `HER_LLM_MODEL`          | `gemma4-e4b`             | Model name to request                    |
| `HER_LLM_ENGLISH_LEVEL`  | `4`                      | Target CEFR level for the conversation \n A1: 1 A2: 2 B1: 3 B2: 4 C1: 5 C2: 6   |
| `HER_SYSTEM_PROMPT`      | *(Samantha persona)*     | Override the system prompt entirely      |

## Development status

Built in phases:

- [x] **Phase 0 — Text loop:** Go + WebSocket skeleton, streaming text chat via a local LLM, core interfaces, unit tests.
- [x] **Phase 1 — Voice in:** push-to-talk mic capture -> 16 kHz WAV -> Gemma 4 (`input_audio`) -> streamed text reply.
- [ ] **Phase 2 — Voice out:** Kokoro TTS (via mlx-audio) + an adjustable speaking-speed / difficulty control.
- [ ] **Phase 3 — Feedback & history:** session-end report (grammar, phrasing, pronunciation) persisted to SQLite.
- [ ] **Phase 4 — Polish:** hands-free VAD, barge-in, sentence-chunked TTS, latency tuning.

## License

[Apache-2.0](LICENSE)

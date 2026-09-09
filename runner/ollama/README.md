# Local model runner

Serves the `local` provider the gateway routes to. Ollama exposes
OpenAI-compatible routes under `/v1`, which is the same wire format the gateway
already speaks to OpenAI and to vLLM, so no gateway code is specific to it.

## Why the binary is installed by hand

Docker Hub is unreachable from Iran, so the usual `ollama/ollama` image is not
an option. `ollama.com` itself answers 403 here, but the model registry at
`registry.ollama.ai` does not, so downloading the release archive once and
pulling models normally both work.

## Use

```bash
./install.sh ~/Downloads/ollama-linux-amd64.tar.zst   # once
./serve.sh                                            # foreground, port 11434
./pull.sh                                             # in another shell
```

`pull.sh` reads the model ids out of `gateway/configs/config.yaml` rather than
keeping its own list, so adding a local model to the catalogue is the only edit
needed.

## Environment

| Variable | Default | Meaning |
|---|---|---|
| `OLLAMA_PREFIX` | `~/.local` | where `bin/ollama` and `lib/ollama` are unpacked |
| `OLLAMA_HOST` | `127.0.0.1:11434` | address the server binds to |
| `GATEWAY_CONFIG` | `gateway/configs/config.yaml` | catalogue `pull.sh` reads |

`OLLAMA_HOST` and `providers.local.base_url` in the gateway config are two ends
of the same connection. Changing one means changing the other.

## Development hardware

The machine this was set up on has an MX550 reporting 1.6 GiB of VRAM with
about 1.1 GiB free, so a 3B model at Q4 does not fit entirely on the GPU and
Ollama splits layers with the CPU. That is a development arrangement only;
production serving is a separate sizing question.

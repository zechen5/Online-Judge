# Online Judge Backend

Backend scaffold for the AI-assisted online judge proposal.

## Run

```bash
go run ./cmd/server
```

Windows teammates can also start everything with one command:

```powershell
.\scripts\start-teammate.ps1
```

Or by double-clicking:

```text
Start-OnlineJudge.bat
```

Windows teammates can stop the local service with:

```powershell
.\scripts\stop-teammate.ps1
```

Or by double-clicking:

```text
Stop-OnlineJudge.bat
```

Linux teammates can start with:

```bash
./scripts/start-teammate.sh
```

And stop with:

```bash
./scripts/stop-teammate.sh
```

macOS teammates can use the same Unix scripts:

```bash
./scripts/start-teammate.sh
./scripts/stop-teammate.sh
```

Defaults:

- `OJ_DB_DRIVER=sqlite`
- `OJ_DB_DSN=app.db`
- `OJ_PORT=8080`

Switch to MySQL by setting `OJ_DB_DRIVER=mysql` and `OJ_DB_DSN`.

## Judge Images

The asynchronous Docker judger needs these local images:

- `algojudge/judge-cpp:latest`
- `algojudge/judge-java:latest`
- `algojudge/judge-python:latest`

Build them locally on Windows with:

```powershell
.\scripts\build-judge-images.ps1
```

Package them for teammates with:

```powershell
.\scripts\package-judge-images.ps1
```

That exports a single archive at:

```text
artifacts\judge-images\algojudge-judge-images.tar
```

Teammates can import the bundle with:

```powershell
.\scripts\load-judge-images.ps1
```

The teammate startup script will try these steps automatically:

1. Check `go` and `docker`.
2. Verify Docker Desktop / Docker Engine is running.
3. Reuse local judge images if they already exist.
4. Otherwise load `artifacts\judge-images\algojudge-judge-images.tar` when present.
5. Otherwise build the judge images from this repository.
6. Start the server and wait for `http://127.0.0.1:8080/healthz`.

On macOS, make sure teammates install:

1. Docker Desktop for Mac
2. Go
3. Xcode Command Line Tools if Docker or Go prompts for them

The stop script exists so teammates can:

1. Free port `8080`.
2. Avoid stale old versions running in the background.
3. Restart cleanly after pulling new code or changing config.

If Docker images are not imported yet, teammates can still run the project as long as they:

1. Install Docker Desktop or Docker Engine.
2. Build the judge images locally from this repository.
3. Start the server with `OJ_JUDGE_EXECUTOR=auto` or `OJ_JUDGE_EXECUTOR=docker`.

# Final year project

## Building

The process for \*NIX oses and Windows differs slightly. For both, Go 1.22+ is required.

The Python build script is compatible with all OSes, but `make` should be preferred for \*NIX-based operating systems.

### \*NIX (Linux and macOS)

To build, use `make`:

```bash
make build          # builds both the server and the game
make run            # runs the server and a single game instance
make run_with_logs  # same as above, but prints the logs to stdout and stderr
```

### Windows

To build, use a Python version 3.10+:

```powershell
python.exe .\build.py build           # builds both the server and the game
python.exe .\build.py run             # runs the server and a single game instance
python.exe .\build.py run_with_logs   # same as above, but prints the logs to stdout and stderr
```

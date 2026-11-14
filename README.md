# crowded

Phones send input; main screens show the combined result. See
[design.md](design.md) for the design.

## Run

```bash
~/.local/go/bin/go run .
```

Then open the addresses it prints:

- Main screen: `http://localhost:8080/crowd`
- Phone page: `http://localhost:8080/`
- Real phone on the same Wi-Fi: the `http://192.168…:8080/` address

Stop with Ctrl-C. Run it again after changing any file.

## Visual experiments

These live in `visual/` and run separately from the backend for now. See
[plan.md](plan.md).

One-time: copy the clip 2 scene from free_range, then install the libraries.

```bash
mkdir -p visual/public/scenes && cp ~/dev/free_range/data/img_gaussian_splat/img.splat visual/public/scenes/clip2.splat
```

```bash
PATH="$HOME/.nvm/versions/node/v22.22.3/bin:$PATH" npm --prefix visual install
```

Run, then open the address it prints (usually `http://localhost:5173/`):

```bash
PATH="$HOME/.nvm/versions/node/v22.22.3/bin:$PATH" npm --prefix visual run dev
```

The `PATH=…` part selects Node 22, which these tools need; this Mac's default
Node is 18. With Node 22 already active, plain `npm --prefix visual run dev`
is enough.

## The short form

Go is installed in `~/.local/go`, and `~/.zshrc` adds it to the PATH, so in
any terminal opened after that change this should work:

```bash
go run .
```

If it says `command not found: go`, the terminal was opened before the
change. Either open a new terminal, or load the change into the current one:

```bash
source ~/.zshrc
```

## After adding or updating a library

```bash
~/.local/go/bin/go mod tidy
```

## First-time setup on another Mac

```bash
mkdir -p ~/.local && curl -L https://go.dev/dl/go1.27.4.darwin-arm64.tar.gz | tar -xz -C ~/.local
```

```bash
echo 'export PATH="$HOME/.local/go/bin:$PATH"' >> ~/.zshrc
```

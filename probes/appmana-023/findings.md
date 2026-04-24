# Phase A probe: appmana-023 (Windows Server 2022, containerd 2.2.1)

Date: 2026-04-24. Cordoned=no (passive read-only probe).

## Named pipe

- `\\.\pipe\containerd-containerd` — gRPC server
- `\\.\pipe\containerd-containerd.ttrpc` — ttrpc
- Listed via `[System.IO.Directory]::GetFiles("\\.\pipe\")` in PowerShell.
- **Gotcha:** `Get-Item "\\.\pipe\containerd-containerd"` FAILS with `ItemNotFound` — named
  pipes aren't accessible via `Get-Item`/`Get-Acl` on PowerShell 5.1. Use the directory
  listing method instead, or open the pipe via `[System.IO.File]::OpenRead`.

## Containerd config (`C:\Program Files\containerd\config.toml`)

- `version = 3` (modern containerd v2 config schema)
- `root = 'C:\ProgramData\containerd\root'`
- `state = 'C:\ProgramData\containerd\state'`
- `[grpc]` `address = '\\.\pipe\containerd-containerd'` — confirms the pipe as the RPC path
- Plugin schema is **NEW** — CRI is split into two plugins:
  - `[plugins.'io.containerd.cri.v1.images']`
    - `discard_unpacked_layers = false` ← ALREADY set (good)
  - `[plugins.'io.containerd.cri.v1.images'.registry]`
    - `config_path = ''` ← MUST set to `C:\ProgramData\containerd\certs.d`
  - `[plugins.'io.containerd.cri.v1.runtime']`
- Legacy `[plugins.'io.containerd.grpc.v1.cri']` entries are still present but have empty
  `config_path` — likely ignored in v3 config. The effective setting is under
  `io.containerd.cri.v1.images.registry`.

## Content store

- `C:\ProgramData\containerd\root\io.containerd.content.v1.content\blobs\sha256` populated
  with real blobs. First 10 samples in `probe.txt`. Sizes range from 1 KiB to 34 MiB.
- ACLs not yet tested (Phase B will enumerate via Go).

## certs.d

- `C:\ProgramData\containerd\certs.d` does NOT exist. Spegel init will create it.

## Snapshotters active

- `io.containerd.snapshotter.v1.cimfs`
- `io.containerd.snapshotter.v1.windows`
- `io.containerd.snapshotter.v1.windows-lcow`

## Implications for the port

1. Windows containerd config drop-in must target **v3 plugin paths**:

   ```toml
   [plugins.'io.containerd.cri.v1.images'.registry]
     config_path = 'C:\ProgramData\containerd\certs.d'
   ```

   The Linux drop-in uses `io.containerd.grpc.v1.cri.registry` — different schema, not
   portable.

2. `discard_unpacked_layers` is already correct on Windows by default — no drop-in change
   needed for that.

3. Spegel's Go client needs to dial `\\.\pipe\containerd-containerd`. containerd/v2 client
   accepts the pipe path directly (not wrapped in `npipe://`); Phase B verifies.

4. Platform default for `ContainerdSock` on Windows: use raw pipe path string, NOT a URI.
   If the go.mod-imported `github.com/containerd/containerd/v2/client` requires
   `winio.DialPipeContext` wrapping, we need a dial helper in `internal/platform/`.
   Phase B answers this.

## Phase B — Go probe results

Built `cmd/probe-containerd-windows` with `GOOS=windows GOARCH=amd64`, ran on appmana-023.

- `client.New("\\\\.\\pipe\\containerd-containerd", client.WithDefaultNamespace("k8s.io"))`
  **works directly**. No `npipe://` scheme, no `winio.DialPipeContext` wrapping. The v2
  client handles the pipe path transparently.
- Version: `v2.2.1` rev `dea7da592f5d1d2b7755e3a161be07f43fad8f75`.
- `ListImages()` returned 259 images (ECR, docker.io, harbor.appmana.com, calico-windows).
- `ContentStore.Walk()` enumerated 1154 blobs totaling 93.6 GB. Metadata is identical shape
  to Linux: digest, size, `Labels` map including `containerd.io/distribution.source.<host>`
  (the exact key Spegel reads for registry-source indexing).
- No Windows-specific Go code needed at the client layer.

## Final platform abstraction scope

Only path defaults + preflight OS gate. No dial-shim. The abstraction is pure configuration:

| Field | Linux default | Windows default |
|---|---|---|
| `ContainerdSock` | `/run/containerd/containerd.sock` | `\\.\pipe\containerd-containerd` |
| `ContainerdContentPath` | `/var/lib/containerd/io.containerd.content.v1.content` | `C:\ProgramData\containerd\root\io.containerd.content.v1.content` |
| `ContainerdRegistryConfigPath` | `/etc/containerd/certs.d` | `C:\ProgramData\containerd\certs.d` |
| `PersistenceHostPath` | `/var/lib/spegel` | `C:\ProgramData\spegel` |
| `BasicAuthSecretsDir` | `/etc/secrets/basic-auth` | `C:\spegel\basic-auth` |

## Windows containerd drop-in (correct format)

For containerd 2.x with config version 3:

```toml
version = 3
[plugins.'io.containerd.cri.v1.images'.registry]
  config_path = 'C:\ProgramData\containerd\certs.d'
```

`discard_unpacked_layers = false` is already the default under
`[plugins.'io.containerd.cri.v1.images']` — no change needed.

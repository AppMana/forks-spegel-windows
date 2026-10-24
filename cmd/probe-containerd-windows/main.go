// Command probe-containerd-windows is a throwaway instrument used during the Windows port
// to verify that the containerd/v2 Go client connects to Windows containerd's named-pipe
// gRPC endpoint and can enumerate the content store the same way it does on Linux.
//
// Not shipped in release artifacts. Lives in the fork to document observed behavior.
//
// Build (from a Linux/Mac workstation):
//
//	GOOS=windows GOARCH=amd64 go build -o probe.exe ./cmd/probe-containerd-windows
//
// Run on a Windows k8s node with containerd 2.x:
//
//	C:\Users\administrator\probe.exe [pipe-path]
//
// Default pipe-path is \\.\pipe\containerd-containerd.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/core/content"
)

func main() {
	os.Exit(run())
}

func run() int {
	sock := `\\.\pipe\containerd-containerd`
	if len(os.Args) > 1 {
		sock = os.Args[1]
	}
	ns := "k8s.io"

	fmt.Fprintf(os.Stdout, "probe: OS=%s ARCH=%s pipe=%s namespace=%s\n", runtime.GOOS, runtime.GOARCH, sock, ns)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	c, err := client.New(sock, client.WithDefaultNamespace(ns))
	if err != nil {
		fmt.Fprintf(os.Stderr, "dial: %v\n", err)
		return 1
	}
	defer c.Close()

	// Versions — confirms the client is actually talking to the server.
	v, err := c.Version(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "version: %v\n", err)
		return 2
	}
	fmt.Fprintf(os.Stdout, "containerd version=%s revision=%s\n", v.Version, v.Revision)

	// Images list — exercises the same ListImages call Spegel uses.
	imgs, err := c.ListImages(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "list images: %v\n", err)
		return 3
	}
	fmt.Fprintf(os.Stdout, "images: %d total\n", len(imgs))
	for i, img := range imgs {
		if i >= 5 {
			fmt.Fprintf(os.Stdout, "... (%d more)\n", len(imgs)-5)
			break
		}
		fmt.Fprintf(os.Stdout, "  %s\n", img.Name())
	}

	// Content store walk — exact call pattern in pkg/oci/containerd.go.
	cs := c.ContentStore()
	count := 0
	var total int64
	if err := cs.Walk(ctx, func(info content.Info) error {
		count++
		total += info.Size
		return nil
	}); err != nil {
		fmt.Fprintf(os.Stderr, "content walk: %v\n", err)
		return 4
	}
	fmt.Fprintf(os.Stdout, "content-store blobs: %d  total=%d bytes\n", count, total)

	// First 5 digest entries for sanity — Spegel uses these as P2P advertisement keys.
	shown := 0
	err = cs.Walk(ctx, func(info content.Info) error {
		if shown >= 5 {
			return io.EOF
		}
		fmt.Fprintf(os.Stdout, "  %s size=%d labels=%v\n", info.Digest, info.Size, info.Labels)
		shown++
		return nil
	})
	if err != nil && !errors.Is(err, io.EOF) {
		fmt.Fprintf(os.Stderr, "content sample: %v\n", err)
		return 5
	}
	return 0
}

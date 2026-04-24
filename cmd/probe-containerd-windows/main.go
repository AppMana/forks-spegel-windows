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
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/core/content"
)

func main() {
	sock := `\\.\pipe\containerd-containerd`
	if len(os.Args) > 1 {
		sock = os.Args[1]
	}
	ns := "k8s.io"

	fmt.Printf("probe: OS=%s ARCH=%s pipe=%s namespace=%s\n", runtime.GOOS, runtime.GOARCH, sock, ns)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	c, err := client.New(sock, client.WithDefaultNamespace(ns))
	if err != nil {
		fmt.Fprintf(os.Stderr, "dial: %v\n", err)
		os.Exit(1)
	}
	defer c.Close()

	// Versions — confirms the client is actually talking to the server.
	v, err := c.Version(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "version: %v\n", err)
		os.Exit(2)
	}
	fmt.Printf("containerd version=%s revision=%s\n", v.Version, v.Revision)

	// Images list — exercises the same ListImages call Spegel uses.
	imgs, err := c.ListImages(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "list images: %v\n", err)
		os.Exit(3)
	}
	fmt.Printf("images: %d total\n", len(imgs))
	for i, img := range imgs {
		if i >= 5 {
			fmt.Printf("... (%d more)\n", len(imgs)-5)
			break
		}
		fmt.Printf("  %s\n", img.Name())
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
		os.Exit(4)
	}
	fmt.Printf("content-store blobs: %d  total=%d bytes\n", count, total)

	// First 5 digest entries for sanity — Spegel uses these as P2P advertisement keys.
	shown := 0
	_ = cs.Walk(ctx, func(info content.Info) error {
		if shown >= 5 {
			return fmt.Errorf("stop")
		}
		fmt.Printf("  %s size=%d labels=%v\n", info.Digest, info.Size, info.Labels)
		shown++
		return nil
	})
}

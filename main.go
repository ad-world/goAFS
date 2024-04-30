package main

import (
	"context"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

var files = map[string]string {
	"file": "contents of file",
	"next_dir/file2": "contents of file 2",
}

type inMemoryFS struct {
	fs.Inode
}

var _ = (fs.NodeOnAdder)((*inMemoryFS)(nil))


// OnAdd is called on mounting the file system. Use it to populate
// the file system tree.
func (root *inMemoryFS) OnAdd(ctx context.Context) {
	for name, content := range files {
		dir, base := filepath.Split(name)

		p := &root.Inode

		// Add directories leading up to the file.
		for _, component := range strings.Split(dir, "/") {
			if len(component) == 0 {
				continue
			}
			ch := p.GetChild(component)
			if ch == nil {
				// Create a directory
				ch = p.NewPersistentInode(ctx, &fs.Inode{},
					fs.StableAttr{Mode: syscall.S_IFDIR})
				// Add it
				p.AddChild(component, ch, true)
			}

			p = ch
		}

		// Make a file out of the content bytes. This type
		// provides the open/read/flush methods.
		embedder := &fs.MemRegularFile{
			Data: []byte(content),
		}

		// Create the file. The Inode must be persistent,
		// because its life time is not under control of the
		// kernel.
		child := p.NewPersistentInode(ctx, embedder, fs.StableAttr{})

		// And add it
		p.AddChild(base, child, true)
	}
}

// This demonstrates how to build a file system in memory. The
// read/write logic for the file is provided by the MemRegularFile type.
func main() {
	// This is where we'll mount the FS
	mntDir, _ := ioutil.TempDir("", "")

	root := &inMemoryFS{}
	server, err := fs.Mount(mntDir, root, &fs.Options{
		MountOptions: fuse.MountOptions{Debug: true},
	})
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Mounted on %s", mntDir)
	log.Printf("Unmount by calling 'fusermount -u %s'", mntDir)

	// Wait until unmount before exiting
	server.Wait()
}

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/caden/agent-recruiting-hub/internal/server"
	"github.com/caden/agent-recruiting-hub/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	root := flag.String("root", "..", "project root (parent of backend/)")
	seedOnStart := flag.Bool("seed", true, "import seed data if db empty")
	flag.Parse()

	absRoot, err := filepath.Abs(*root)
	if err != nil {
		log.Fatal(err)
	}

	st, err := store.Open(absRoot)
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	srv := server.New(st, absRoot)

	if *seedOnStart {
		n, _ := st.Count()
		if n == 0 {
			if count, err := srv.RunSeedImport(); err != nil {
				log.Printf("seed import failed: %v", err)
			} else {
				log.Printf("auto-seeded %d candidates", count)
			}
		}
	}

	log.Printf("agent-recruiting-hub listening on http://localhost%s", *addr)
	log.Printf("data dir: %s", filepath.Join(absRoot, "data"))
	if err := srv.Router().Run(*addr); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

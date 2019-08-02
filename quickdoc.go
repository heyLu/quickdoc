package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	stdlibPkgs, err := ListStdlibPkgs()
	if err != nil {
		log.Fatalf("listing go stdlib packages: %s", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, `<!doctype html>
<html>
<head>
	<meta charset="utf-8" />
	<title>quickdoc</title>
	<style>
	body {
		margin: 0;
	}

	#result {
		border: none;
		width: 100vw;
		height: 100vh;
	}
	</style>
</head>

<body>
	<h1>quickdoc</h1>

	<input id="search" type="search" autofocus />
	<hr />

	<iframe id="result" src="/doc"></iframe>

	<script>
		var searchEl = document.querySelector("#search");
		var resultEl = document.querySelector("#result");

		// init if set via browser cache
		resultEl.src = "/doc/" + searchEl.value;

		var timeout = null;
		searchEl.addEventListener("input", function(ev) {
			clearTimeout(timeout);

			timeout = setTimeout(function() {
				resultEl.src = "/doc/" + searchEl.value;
			}, 200);
		});
	</script>
</body>
</html>`)
	})

	http.HandleFunc("/doc/", func(w http.ResponseWriter, req *http.Request) {
		pkgName := req.URL.Path[len("/doc/"):]

		if pkgName == "" {
			fmt.Fprintf(w, `Usage:

/net -> renders go/net docs
/net ListenIP -> renders ListenIP docs in net
/net.ListenIP
/net/http -> renders net/http docs

write -> renders man page for write
2 write -> renders man page for write(2)

!h ag -> renders ag --help
ag!h
			`)
			return
		}

		switch {
		case IsStdlibPkg(stdlibPkgs, pkgName):
			err := RenderGoDoc(w, pkgName)
			if err != nil {
				log.Printf("could not render godoc: %s", err)
				fmt.Fprintf(w, "\n\nERROR\n")
				return
			}
		case strings.Contains(pkgName, "!h"):
			cmdName := strings.TrimSpace(strings.Replace(pkgName, "!h", "", 1))
			err := RenderProgramHelp(w, cmdName)
			if err != nil {
				log.Printf("could not render help for %s: %s", cmdName, err)
				fmt.Fprintf(w, "\n\nERROR\n")
				return
			}
		default:
			err := RenderUnixMan(w, pkgName)
			if err != nil {
				log.Printf("could not render man page: %s", err)
				fmt.Fprintf(w, "\n\nERROR\n")
				return
			}
		}
	})

	addr := "localhost:9998"
	log.Printf("listening on http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func RenderGoDoc(w io.Writer, pkgName string) error {
	cmd := exec.Command("go", "doc", pkgName)
	cmd.Stdout = w
	cmd.Stderr = w
	return cmd.Run()
}

func RenderUnixMan(w io.Writer, manPage string) error {
	cmd := exec.Command("man", strings.Split(manPage, " ")...)
	cmd.Stdout = w
	cmd.Stderr = w
	return cmd.Run()
}

func RenderProgramHelp(w io.Writer, cmdName string) error {
	cmd := exec.Command(cmdName, "--help")
	cmd.Stdout = w
	cmd.Stderr = w
	return cmd.Run()
}

func IsStdlibPkg(stdlibPkgs []string, pkgSpec string) bool {
	onlyPkgName := pkgSpec
	specifierStart := strings.IndexAny(pkgSpec, ".")
	if specifierStart != -1 {
		onlyPkgName = onlyPkgName[:specifierStart]
	}

	for _, pkg := range stdlibPkgs {
		if pkg == onlyPkgName {
			return true
		}
	}

	return false
}

func ListStdlibPkgs() ([]string, error) {
	pkgRoot := filepath.Join(runtime.GOROOT(), "pkg", runtime.GOOS+"_"+runtime.GOARCH)
	pkgs := make([]string, 0, 30)
	err := filepath.Walk(pkgRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == pkgRoot {
			return nil
		}
		base := filepath.Base(path)
		if base == "vendor" || base == "cmd" || base == "internal" {
			return filepath.SkipDir
		}

		ext := filepath.Ext(path)
		if ext != "" {
			path = path[:len(path)-len(ext)]
		}

		pkgs = append(pkgs, path[len(pkgRoot)+1:])

		return nil
	})
	return pkgs, err
}

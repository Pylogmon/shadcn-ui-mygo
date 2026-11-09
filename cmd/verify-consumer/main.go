// verify-consumer packages a committed revision into a local Go module proxy,
// then builds an independent application and runs the Gallery from that proxy.
// All child builds use initially empty caches and no workspace or replacements.
package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"image/png"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const kitModule = "github.com/pylogmon/shadcn-ui-mygo"

type moduleInfo struct {
	Path, Version, Dir string
	Replace            *moduleInfo
}

type receipt struct {
	Commit, Version, GoVersion, HostOS, HostArch string
	ModuleZIP, ConsumerSource                    string
	ModuleZIP_SHA256, ConsumerSource_SHA256      string
	IsolatedModuleCache, IsolatedBuildCache      bool
	WorkspaceDisabled, NoModuleReplacements      bool
	Modules                                      []moduleInfo
}

func main() {
	ref := flag.String("ref", "HEAD", "committed SDK revision to package")
	out := flag.String("out", "build/consumer-verification", "parent directory for retained evidence")
	flag.Parse()
	if err := verify(*ref, *out); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func output(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	b, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, b)
	}
	return b, nil
}

func verify(ref, parent string) error {
	rootBytes, err := output("", "rev-parse", "--show-toplevel")
	if err != nil {
		return err
	}
	root := strings.TrimSpace(string(rootBytes))
	commitBytes, err := output(root, "rev-parse", "--verify", "--end-of-options", ref+"^{commit}")
	if err != nil {
		return err
	}
	commit := strings.TrimSpace(string(commitBytes))
	if len(commit) != 40 {
		return fmt.Errorf("expected a SHA-1 commit, got %q", commit)
	}
	if _, err := hex.DecodeString(commit); err != nil {
		return err
	}
	dateBytes, err := output(root, "show", "-s", "--format=%cI", commit)
	if err != nil {
		return err
	}
	date, err := time.Parse(time.RFC3339, strings.TrimSpace(string(dateBytes)))
	if err != nil {
		return err
	}
	version := "v0.0.0-" + date.UTC().Format("20060102150405") + "-" + commit[:12]
	if !filepath.IsAbs(parent) {
		parent = filepath.Join(root, parent)
	}
	if err := os.MkdirAll(parent, 0755); err != nil {
		return err
	}
	dir, err := os.MkdirTemp(parent, "run-")
	if err != nil {
		return err
	}
	fmt.Println("Consumer evidence:", dir)

	cache, err := os.MkdirTemp("", "shadcn-consumer-caches-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(cache)
	proxy := filepath.Join(dir, "proxy")
	versions := filepath.Join(proxy, filepath.FromSlash(kitModule), "@v")
	if err := os.MkdirAll(versions, 0755); err != nil {
		return err
	}
	zipPath := filepath.Join(versions, version+".zip")
	archive, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	cmd := exec.Command("git", "archive", "--format=zip", "--prefix="+kitModule+"@"+version+"/", commit)
	cmd.Dir, cmd.Stdout, cmd.Stderr = root, archive, os.Stderr
	archiveErr := cmd.Run()
	closeErr := archive.Close()
	if archiveErr != nil {
		return archiveErr
	}
	if closeErr != nil {
		return closeErr
	}
	archiveReader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("read SDK archive: %w", err)
	}
	modulePrefix := kitModule + "@" + version + "/"
	hasModule := false
	for _, file := range archiveReader.File {
		if !strings.HasPrefix(file.Name, modulePrefix) {
			archiveReader.Close()
			return fmt.Errorf("file outside module archive prefix: %s", file.Name)
		}
		if file.Name == modulePrefix+"go.mod" {
			hasModule = true
		}
	}
	if err := archiveReader.Close(); err != nil {
		return err
	}
	if !hasModule {
		return fmt.Errorf("SDK module archive has no go.mod")
	}
	mod, err := output(root, "show", commit+":go.mod")
	if err != nil {
		return err
	}
	if !bytes.HasPrefix(mod, []byte("module "+kitModule+"\n")) {
		return fmt.Errorf("unexpected source module declaration")
	}
	if err := os.WriteFile(filepath.Join(versions, version+".mod"), mod, 0644); err != nil {
		return err
	}
	info, _ := json.Marshal(map[string]string{"Version": version, "Time": date.UTC().Format(time.RFC3339)})
	if err := os.WriteFile(filepath.Join(versions, version+".info"), info, 0644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(versions, "list"), []byte(version+"\n"), 0644); err != nil {
		return err
	}

	// This fixture is deliberately outside the SDK artifact. It is a consumer
	// of the selected revision, and its exact bytes are recorded independently.
	fixture, err := os.ReadFile(filepath.Join(root, "examples", "consumer", "main.go"))
	if err != nil {
		return err
	}
	consumer := filepath.Join(dir, "consumer")
	if err := os.MkdirAll(consumer, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(consumer, "main.go"), fixture, 0644); err != nil {
		return err
	}
	goVersion := ""
	for _, line := range strings.Split(string(mod), "\n") {
		if strings.HasPrefix(line, "go ") {
			goVersion = strings.TrimSpace(strings.TrimPrefix(line, "go "))
			break
		}
	}
	if goVersion == "" {
		return fmt.Errorf("SDK go directive missing")
	}
	consumerMod := "module example.invalid/native-consumer\n\ngo " + goVersion + "\n\nrequire " + kitModule + " " + version + "\n"
	if err := os.WriteFile(filepath.Join(consumer, "go.mod"), []byte(consumerMod), 0644); err != nil {
		return err
	}
	env := isolatedEnv(map[string]string{
		"GOENV": "off", "GOWORK": "off", "GOTOOLCHAIN": "local", "GOFLAGS": "", "CGO_ENABLED": "0",
		"GOMODCACHE": filepath.Join(cache, "modules"), "GOCACHE": filepath.Join(cache, "build"),
		"GOPROXY":   (&url.URL{Scheme: "file", Path: filepath.ToSlash(proxy)}).String() + ",https://proxy.golang.org",
		"GONOPROXY": "none", "GOPRIVATE": "", "GONOSUMDB": kitModule, "GOSUMDB": "sum.golang.org",
	})
	run := func(name string, args ...string) ([]byte, error) {
		fmt.Println("Verifying:", name)
		cmd := exec.Command("go", args...)
		cmd.Dir, cmd.Env = consumer, env
		b, err := cmd.CombinedOutput()
		if writeErr := os.WriteFile(filepath.Join(dir, name+".log"), b, 0644); writeErr != nil {
			return b, writeErr
		}
		if err != nil {
			return b, fmt.Errorf("%s failed: %w\n%s", name, err, b)
		}
		return b, nil
	}
	if _, err := run("tidy", "mod", "tidy"); err != nil {
		return err
	}
	modulesJSON, err := run("modules", "list", "-m", "-json", "all")
	if err != nil {
		return err
	}
	var modules []moduleInfo
	decoder := json.NewDecoder(bytes.NewReader(modulesJSON))
	found := false
	for {
		var module moduleInfo
		if err := decoder.Decode(&module); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if module.Replace != nil {
			return fmt.Errorf("consumer graph contains replacement for %s", module.Path)
		}
		if module.Path == kitModule {
			found = true
			if module.Version != version || !strings.HasPrefix(module.Dir, filepath.Join(cache, "modules")+string(os.PathSeparator)) {
				return fmt.Errorf("SDK did not resolve from isolated module cache: %+v", module)
			}
		}
		modules = append(modules, module)
	}
	if !found {
		return fmt.Errorf("SDK missing from consumer module graph")
	}
	if _, err := run("build", "build", "-o", filepath.Join(dir, "consumer-app"), "."); err != nil {
		return err
	}
	if _, err := run("consumer-render", "run", ".", "-snapshot", filepath.Join(dir, "consumer.png"), "-preferences", filepath.Join(dir, "theme.json"), "-verify-theme"); err != nil {
		return err
	}
	if _, err := run("consumer-restore", "run", ".", "-snapshot", filepath.Join(dir, "consumer-restored.png"), "-preferences", filepath.Join(dir, "theme.json")); err != nil {
		return err
	}
	if _, err := run("gallery-light", "run", kitModule+"/cmd/gallery@"+version, "-snapshot", filepath.Join(dir, "gallery-light.png")); err != nil {
		return err
	}
	if _, err := run("gallery-dark", "run", kitModule+"/cmd/gallery@"+version, "-dark", "-snapshot", filepath.Join(dir, "gallery-dark.png")); err != nil {
		return err
	}
	for _, name := range []string{"consumer.png", "consumer-restored.png", "gallery-light.png", "gallery-dark.png"} {
		f, err := os.Open(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		config, decodeErr := png.DecodeConfig(f)
		f.Close()
		if decodeErr != nil || config.Width <= 0 || config.Height <= 0 {
			return fmt.Errorf("invalid rendered %s: %v", name, decodeErr)
		}
	}
	// A second process loads the settings without invoking the verifier's Save
	// path. Compare the actual theme and scoped controls above the status line.
	if err := compareRestoredTheme(filepath.Join(dir, "consumer.png"), filepath.Join(dir, "consumer-restored.png")); err != nil {
		return err
	}
	goEnv, err := run("go-env", "env", "-json", "GOVERSION", "GOOS", "GOARCH", "GOWORK", "GOMODCACHE", "GOCACHE")
	if err != nil {
		return err
	}
	var effective map[string]string
	if err := json.Unmarshal(goEnv, &effective); err != nil {
		return err
	}
	zipHash, err := fileHash(zipPath)
	if err != nil {
		return err
	}
	fixtureHash := sha256.Sum256(fixture)
	r := receipt{Commit: commit, Version: version, GoVersion: effective["GOVERSION"], HostOS: effective["GOOS"], HostArch: effective["GOARCH"],
		ModuleZIP: zipPath, ConsumerSource: filepath.Join(consumer, "main.go"), ModuleZIP_SHA256: zipHash, ConsumerSource_SHA256: hex.EncodeToString(fixtureHash[:]),
		IsolatedModuleCache: true, IsolatedBuildCache: true, WorkspaceDisabled: effective["GOWORK"] == "off", NoModuleReplacements: true, Modules: modules}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "receipt.json"), append(data, '\n'), 0644); err != nil {
		return err
	}
	fmt.Println("Verified committed SDK:", commit, version)
	return nil
}

func isolatedEnv(overrides map[string]string) []string {
	env := make([]string, 0, len(os.Environ())+len(overrides))
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if _, replaced := overrides[key]; !replaced {
			env = append(env, entry)
		}
	}
	for key, value := range overrides {
		env = append(env, key+"="+value)
	}
	return env
}

func fileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func compareRestoredTheme(first, second string) error {
	a, err := os.Open(first)
	if err != nil {
		return err
	}
	defer a.Close()
	b, err := os.Open(second)
	if err != nil {
		return err
	}
	defer b.Close()
	before, err := png.Decode(a)
	if err != nil {
		return err
	}
	after, err := png.Decode(b)
	if err != nil {
		return err
	}
	if before.Bounds() != after.Bounds() || before.Bounds().Dx() != 560 || before.Bounds().Dy() != 520 {
		return fmt.Errorf("unexpected consumer snapshot dimensions")
	}
	for y := 0; y < 400; y++ {
		for x := 0; x < 560; x++ {
			r, g, b, a := before.At(x, y).RGBA()
			r2, g2, b2, a2 := after.At(x, y).RGBA()
			if r != r2 || g != g2 || b != b2 || a != a2 {
				return fmt.Errorf("fresh consumer process did not restore theme pixels at (%d,%d)", x, y)
			}
		}
	}
	return nil
}

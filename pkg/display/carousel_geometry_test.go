package display

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

const carouselSourceArchive = "testdata/shadcn-carousel"

func carouselArchiveRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(carouselSourceArchive, path))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestCarouselPinnedSourceAndOracleIntegrity(t *testing.T) {
	var manifest struct {
		Commit string
		Files  []struct{ Path, Archive, GitBlob, SHA256 string }
		NPM    []struct {
			Name, Version, Integrity, Archive, TarballSHA256 string
			Files                                            []struct{ Path, Archive, SHA256 string }
		}
		EmblaLicense struct{ Archive, SHA256 string }
		Oracle       struct{ Archive, SHA256, Generator, GeneratorSHA256 string }
	}
	if err := json.Unmarshal(carouselArchiveRead(t, "manifest.json"), &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Commit != "8ab27715521cbd4f87ae6886cd989cbb678f6de7" || len(manifest.Files) != 15 || len(manifest.NPM) != 4 {
		t.Fatal("source scope/pin changed without updating the parity baseline")
	}
	checkSHA := func(path, want string) []byte {
		t.Helper()
		data := carouselArchiveRead(t, path)
		got := sha256.Sum256(data)
		if hex.EncodeToString(got[:]) != want {
			t.Fatalf("%s SHA256 mismatch", path)
		}
		return data
	}
	for _, source := range manifest.Files {
		data := checkSHA(source.Archive, source.SHA256)
		blob := sha1.New()
		fmt.Fprintf(blob, "blob %d\x00", len(data))
		blob.Write(data)
		if hex.EncodeToString(blob.Sum(nil)) != source.GitBlob {
			t.Fatalf("%s does not match pinned Git blob", source.Path)
		}
	}
	lock := carouselArchiveRead(t, "shadcn/pnpm-lock.yaml")
	for _, pkg := range manifest.NPM {
		if pkg.Version != "8.5.2" || !bytes.Contains(lock, []byte(pkg.Integrity)) {
			t.Fatalf("%s is not from the pinned lock", pkg.Name)
		}
		data := checkSHA(pkg.Archive, pkg.TarballSHA256)
		integrity := sha512.Sum512(data)
		if "sha512-"+base64.StdEncoding.EncodeToString(integrity[:]) != pkg.Integrity {
			t.Fatalf("%s npm tarball integrity mismatch", pkg.Name)
		}
		gz, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		members := make(map[string][]byte)
		reader := tar.NewReader(gz)
		for {
			header, err := reader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			if header.Typeflag == tar.TypeReg {
				member, err := io.ReadAll(reader)
				if err != nil {
					t.Fatal(err)
				}
				members[header.Name] = member
			}
		}
		gz.Close()
		for _, source := range pkg.Files {
			if !bytes.Equal(checkSHA(source.Archive, source.SHA256), members[source.Path]) {
				t.Fatalf("%s differs from verified npm member", source.Archive)
			}
		}
	}
	checkSHA(manifest.EmblaLicense.Archive, manifest.EmblaLicense.SHA256)
	checkSHA(manifest.Oracle.Archive, manifest.Oracle.SHA256)
	checkSHA(manifest.Oracle.Generator, manifest.Oracle.GeneratorSHA256)
}

func TestCarouselGeometryMatchesIndependentEmblaOracle(t *testing.T) {
	var fixture struct {
		Source, SourceSHA256 string
		CaseCount            int
		Cases                []struct {
			Name, Align, Contain    string
			Viewport, Spacing, View float32
			Total                   float32
			Bases                   []struct{ Fraction, Fixed float32 }
			RequestedLoop, Loop     bool
			Sizes, Starts, Snaps    []float32
			Poses                   []struct {
				Snap  float32
				Items []struct {
					Index                int
					OuterStart, OuterEnd float32
				}
			}
		}
	}
	if err := json.Unmarshal(carouselArchiveRead(t, "geometry-oracle.json"), &fixture); err != nil {
		t.Fatal(err)
	}
	source := sha256.Sum256(carouselArchiveRead(t, fixture.Source))
	if hex.EncodeToString(source[:]) != fixture.SourceSHA256 || fixture.CaseCount != 324 || len(fixture.Cases) != fixture.CaseCount {
		t.Fatal("independent oracle source/case count changed")
	}
	for caseIndex, input := range fixture.Cases {
		name := strings.Join([]string{strconv.Itoa(caseIndex), input.Name, input.Align, input.Contain, fmt.Sprint(input.Spacing, input.RequestedLoop)}, "/")
		align := map[string]CarouselAlign{"start": CarouselAlignStart, "center": CarouselAlignCenter, "end": CarouselAlignEnd}[input.Align]
		contain := map[string]CarouselContain{"trimSnaps": CarouselTrimSnaps, "keepSnaps": CarouselKeepSnaps, "none": CarouselUncontained}[input.Contain]
		bases := make([]CarouselBasis, len(input.Bases))
		for i, basis := range input.Bases {
			if basis.Fixed > 0 {
				bases[i] = CarouselFixed(basis.Fixed)
			} else {
				bases[i] = CarouselFraction(basis.Fraction)
			}
		}
		actual := carouselMeasure(input.Viewport, input.Spacing, bases, align, contain, input.RequestedLoop)
		compare := func(label string, got, want []float32) {
			t.Helper()
			if len(got) != len(want) {
				t.Fatalf("%s %s length: got=%v want=%v", name, label, got, want)
			}
			for i := range got {
				if carouselAbs(got[i]-want[i]) > .002 {
					t.Fatalf("%s %s[%d]: got=%v want=%v", name, label, i, got, want)
				}
			}
		}
		compare("sizes", actual.sizes, input.Sizes)
		compare("starts", actual.starts, input.Starts)
		compare("snaps", actual.snaps, input.Snaps)
		compare("view/total", []float32{actual.view, actual.total}, []float32{input.View, input.Total})
		if actual.loop != input.Loop {
			t.Fatalf("%s resolved loop=%v want %v", name, actual.loop, input.Loop)
		}
		for snapIndex, pose := range input.Poses {
			position := actual.position(snapIndex)
			var visible, expected []int
			for i, item := range pose.Items {
				shown := item.OuterStart < input.Viewport-.01 && item.OuterEnd > .01
				if shown {
					expected = append(expected, i)
				}
				if actual.visible(i, position) {
					visible = append(visible, i)
					start := actual.starts[i] - actual.spacing + position + actual.occurrence(i, position)
					compare("visible outer bounds", []float32{start, start + actual.sizes[i]}, []float32{item.OuterStart, item.OuterEnd})
				}
			}
			if !slices.Equal(visible, expected) {
				t.Fatalf("%s snap%d visible=%v want=%v", name, snapIndex, visible, expected)
			}
		}
	}
}

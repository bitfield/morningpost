package morningpost_test

import (
	"errors"
	"io/fs"
	"os"
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/thiagonache/morningpost"
)

func newFileStoreWithBogusPath(t *testing.T) *morningpost.FileStore {
	fileStore, err := morningpost.NewFileStore(
		morningpost.WithFileStorePath(t.TempDir() + "/bogus"),
	)
	if err != nil {
		t.Fatal(err)
	}
	return fileStore
}
func TestAdd_PopulatesStoreGivenFeed(t *testing.T) {
	want := map[string]morningpost.Feed{
		"776ae47727e1e4a3c3b41761e514b146a508f487908574ceb577aa1c2ac12dad": {
			Endpoint: "fake.url",
			ID:       "776ae47727e1e4a3c3b41761e514b146a508f487908574ceb577aa1c2ac12dad",
		},
	}
	fileStore := newFileStoreWithBogusPath(t)
	fileStore.Add(morningpost.Feed{
		Endpoint: "fake.url",
	})
	got := fileStore.Data
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestGetAll_ReturnsProperItemsGivenPrePoluatedStore(t *testing.T) {
	t.Parallel()
	want := []morningpost.Feed{
		{
			Endpoint: "http://fake-http.url",
		},
		{
			Endpoint: "https://fake-https.url",
		},
	}
	fileStore := newFileStoreWithBogusPath(t)
	fileStore.Data = map[string]morningpost.Feed{
		"http://fake-http.url": {
			Endpoint: "http://fake-http.url",
		},
		"https://fake-https.url": {
			Endpoint: "https://fake-https.url",
		},
	}
	got := fileStore.GetAll()
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestLoad_ReturnsExpectedDataGivenEmptyFileStore(t *testing.T) {
	t.Parallel()
	want := map[string]morningpost.Feed{}
	fileStore := newFileStoreWithBogusPath(t)
	err := fileStore.Load()
	if err != nil {
		t.Fatal(err)
	}
	got := fileStore.Data
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestSave_PersistsDataToStore(t *testing.T) {
	t.Parallel()
	want := map[string]morningpost.Feed{
		"http://fake-http.url": {
			Endpoint: "http://fake-http.url",
		},
		"https://fake-https.url": {
			Endpoint: "https://fake-https.url",
		},
	}
	tempDir := t.TempDir()
	fileStore, err := morningpost.NewFileStore(
		morningpost.WithFileStorePath(tempDir + "/bogus"),
	)
	if err != nil {
		t.Fatal(err)
	}
	fileStore.Data = map[string]morningpost.Feed{
		"http://fake-http.url": {
			Endpoint: "http://fake-http.url",
		},
		"https://fake-https.url": {
			Endpoint: "https://fake-https.url",
		},
	}
	err = fileStore.Save()
	if err != nil {
		t.Fatal(err)
	}
	fileStore2, err := morningpost.NewFileStore(
		morningpost.WithFileStorePath(tempDir + "/bogus"),
	)
	if err != nil {
		t.Fatal(err)
	}
	got := fileStore2.Data
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestDelete_RemovesFeedFromStore(t *testing.T) {
	t.Parallel()
	want := map[string]morningpost.Feed{
		"https://fake-https.url": {
			Endpoint: "https://fake-https.url",
		},
	}
	fileStore := newFileStoreWithBogusPath(t)
	fileStore.Data = map[string]morningpost.Feed{
		"http://fake-http.url": {
			Endpoint: "http://fake-http.url",
		},
		"https://fake-https.url": {
			Endpoint: "https://fake-https.url",
		},
	}
	fileStore.Delete("http://fake-http.url")
	got := fileStore.Data
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestNewFileStore_SetsDefaultPath(t *testing.T) {
	t.Parallel()
	fileStorePaths := map[string]string{
		"darwin": os.ExpandEnv("$HOME") + "/Library/Application Support/MorningPost/morningpost.db",
		"linux":  "/var/lib/MorningPost/morningpost.db",
	}
	want := fileStorePaths[runtime.GOOS]
	fileStore, err := morningpost.NewFileStore()
	if err != nil {
		t.Fatal(err)
	}
	got := fileStore.Path
	if want != got {
		t.Fatalf("want Path %q, got %q", want, got)
	}
}

func TestNewFileStore_LoadsDataGivenPopulatedStore(t *testing.T) {
	t.Parallel()
	want := map[string]morningpost.Feed{
		"http://fake-http.url": {
			Endpoint: "http://fake-http.url",
		},
		"https://fake-https.url": {
			Endpoint: "https://fake-https.url",
		},
	}
	fileStore, err := morningpost.NewFileStore(
		morningpost.WithFileStorePath("testdata/golden/filestore.gob"),
	)
	if err != nil {
		t.Fatal(err)
	}
	got := fileStore.Data
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestNewFileStore_CreatesDirectoryGivenPathNotExist(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	_, err := morningpost.NewFileStore(
		morningpost.WithFileStorePath(tempDir + "/directory/bogus/file.db"),
	)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(tempDir + "/directory/bogus")
	if errors.Is(err, os.ErrNotExist) {
		t.Fatalf("want to exist but it doesn't")
	}
	if !info.IsDir() {
		t.Fatalf("want %q to be a directory but it is not", tempDir+"/directory/bogus")
	}
	if info.Mode().Perm() != fs.FileMode(0755) {
		t.Fatalf("want permission %v, got %v", fs.FileMode(0755), info.Mode().Perm())
	}

}

func TestWithFileStorePath_SetsPathGivenString(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	want := tempDir + "/morningpost.db"
	fileStore, err := morningpost.NewFileStore(
		morningpost.WithFileStorePath(tempDir + "/morningpost.db"),
	)
	if err != nil {
		t.Fatal(err)
	}
	got := fileStore.Path
	if want != got {
		t.Fatalf("want Path %q, got %q", want, got)
	}
}

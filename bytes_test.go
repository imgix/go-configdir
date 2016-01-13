package configdir

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func init() {
	logger = log.New(os.Stderr, "[HwyServer-Worker]",
		log.Ldate|log.Ltime|log.Lshortfile,
	)

}

func writeManyFiles(t *testing.T, d, suffix string, numfiles int, body []byte) []byte {
	var (
		fname string
		buf   *bytes.Buffer
	)
	buf = &bytes.Buffer{}
	for x := 0; x < numfiles; x++ {
		bufBody := bytes.NewBuffer(body)

		// gotta zero pad so dirctory sorting returns the same
		fname = filepath.Join(d, fmt.Sprintf("test_%06d.%s", x, suffix))

		bufBody.Write([]byte(fname))
		if err := ioutil.WriteFile(fname, bufBody.Bytes(), os.ModePerm); err != nil {
			t.Fatal(err)
		}
		buf.Write(bufBody.Bytes())

	}
	return buf.Bytes()

}

func cleanupDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer os.Remove(dir)
	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func TestBytesFromDir(t *testing.T) {
	d, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)

	}
	defer cleanupDir(d)
	wroteBytes := writeManyFiles(t, d, "toml", 20, []byte("dir files data"))
	readBytes, readmd5 := bytesFromDir(d, "toml")

	wrotemd5 := md5.New()
	wrotemd5.Write(wroteBytes)

	if !bytes.Equal(wrotemd5.Sum(nil), readmd5) {

		msg := fmt.Sprintf("wrote bytes: \n%s\nread bytes:\n%s", wroteBytes, readBytes)
		t.Error(msg)
	}

}

//
func TestDirectoryUpdates(t *testing.T) {
	d, err := ioutil.TempDir("", "")

	if err != nil {
		t.Fatal(err)

	}
	defer cleanupDir(d)
	writeManyFiles(t, d, "toml", 20, []byte("initial"))

	chBytes, err := DirectoryUpdates(d, "toml", nil)
	if err != nil {
		t.Fatal(err)
	}

	<-chBytes

	for x := 0; x < 5; x++ {
		fname := filepath.Join(d, fmt.Sprintf("test_%06d.toml", x))
		if err := ioutil.WriteFile(fname, []byte("foo"), os.ModePerm); err != nil {
			t.Fatal(err)
		}

		b := <-chBytes
		bfromDir, _ := bytesFromDir(d, "toml")

		if !bytes.Equal(b, bfromDir) {
			t.Errorf("directory update returns inconsistent data")
		}

	}

	// test no-op
	fname := filepath.Join(d, fmt.Sprintf("test_%06d.toml", 6))
	if err := ioutil.WriteFile(fname, []byte("foo"), os.ModePerm); err != nil {
		t.Fatal(err)
	}
	select {

	case <-chBytes:
		t.Errorf("md5 check let no-op file write slip through")
	default:

	}

}
func makeToml(s string) string {
	return s
}

func ExampleDirectoryUpdates() {
	logger = log.New(os.Stderr, "[dircfg]",
		log.Ldate|log.Ltime|log.Lshortfile,
	)

	chBytes, err := DirectoryUpdates("/etc/server.d", "toml", logger)
	if err != nil {
		panic(err)
	}
	var b []byte
	b = <-chBytes

	// dummy config parser func
	makeToml(string(b))
}

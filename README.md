# go-configdir: channel upates for config directories

[![GODOC](https://godoc.org/github.com/imgix/go-confidir?status.svg)](http://godoc.org/github.com/imgix/go-configdir)

receive channel-based updates as when directory contents are modified.

a go-routine reads all files into a single bytes.Buffer and compares against previous reads to only notify on unique changes.

this package should be limited to config file updates or other directory contents that fit reasonably in memory.

```golang


	chBytes, err := DirectoryUpdates("/etc/server.d", "toml", logger)
	if err != nil {
		panic(err)
	}
	var buf *bytes.Buffer
	// uniquified byte buffers arrive from channel
	// directory files are concatenated and md5 sum checked
	buf = <-chBytes

	// dummy config parser func
	makeToml(buf.String())
```

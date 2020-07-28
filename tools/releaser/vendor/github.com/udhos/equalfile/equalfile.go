package equalfile

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
)

// Only the first 10^10 bytes of io.Reader are compared.  Ignored when using io.LimitedReader
const defaultMaxSize = 10000000000
const defaultBufSize = 20000

type Options struct {
	Debug         bool // enable debugging to stdout
	ForceFileRead bool // prevent shortcut at filesystem level (link, pathname, etc)

	// MaxSize is a safely limit to prevent forever reading from an infinite
	// reader.  If left unset, will default to 1OGBytes. Ignored when
	// CompareReader() is given one or more io.LimitedReader.
	MaxSize int64
}

type Cmp struct {
	Opt Options

	readCount int
	readMin   int
	readMax   int
	readSum   int64

	hashType         hash.Hash
	hashMatchCompare bool
	hashTable        map[string]hashSum

	buf []byte
}

type hashSum struct {
	result []byte
	err    error
}

// New creates Cmp for multiple comparison mode.
func NewMultiple(buf []byte, options Options, h hash.Hash, compareOnMatch bool) *Cmp {
	c := &Cmp{
		Opt:              options,
		hashType:         h,
		hashMatchCompare: compareOnMatch,
		hashTable:        map[string]hashSum{},
		buf:              buf,
	}
	if c.buf == nil || len(c.buf) == 0 {
		c.buf = make([]byte, defaultBufSize)
	}
	c.debugf("New: bufSize=%d\n", len(c.buf))
	return c
}

// New creates Cmp for single comparison mode.
func New(buf []byte, options Options) *Cmp {
	return NewMultiple(buf, options, nil, true)
}

func (c *Cmp) getHash(path string, maxSize int64) ([]byte, error) {
	h, found := c.hashTable[path]
	if found {
		return h.result, h.err
	}

	f, openErr := os.Open(path)
	if openErr != nil {
		return nil, openErr
	}
	defer f.Close()

	sum := make([]byte, c.hashType.Size())
	c.hashType.Reset()
	n, copyErr := io.CopyN(c.hashType, f, maxSize)
	copy(sum, c.hashType.Sum(nil))

	if copyErr == io.EOF && n < maxSize {
		copyErr = nil
	}

	return c.newHash(path, sum, copyErr)
}

func (c *Cmp) newHash(path string, sum []byte, e error) ([]byte, error) {

	c.hashTable[path] = hashSum{sum, e}

	c.debugf("newHash[%s]=%v: error=[%v]\n", path, hex.EncodeToString(sum), e)

	return sum, e
}

func (c *Cmp) multipleMode() bool {
	return c.hashType != nil
}

// CompareFile verifies that files with names path1, path2 have same contents.
func (c *Cmp) CompareFile(path1, path2 string) (bool, error) {

	if c.Opt.MaxSize < 0 {
		return false, fmt.Errorf("negative MaxSize")
	}

	r1, openErr1 := os.Open(path1)
	if openErr1 != nil {
		return false, openErr1
	}
	defer r1.Close()
	info1, statErr1 := r1.Stat()
	if statErr1 != nil {
		return false, statErr1
	}

	r2, openErr2 := os.Open(path2)
	if openErr2 != nil {
		return false, openErr2
	}
	defer r2.Close()
	info2, statErr2 := r2.Stat()
	if statErr2 != nil {
		return false, statErr2
	}

	if !c.Opt.ForceFileRead {
		// shortcut: ask the filesystem: are these files the same? (link, pathname, etc)
		if os.SameFile(info1, info2) {
			c.debugf("CompareFile(%s,%s): os reported same file\n", path1, path2)
			return true, nil
		}
	}

	if info1.Mode().IsRegular() && info2.Mode().IsRegular() {
		if info1.Size() != info2.Size() {
			c.debugf("CompareFile(%s,%s): distinct file sizes\n", path1, path2)
			return false, nil
		}
	}

	// If Opt.MaxSize not initialized, set maxSize to an appropriate value
	// for comparing regular files or streams (pipes, devices), etc.
	// Pass maxSize to getHash to ensure the hash is computed only up to
	// the user specified MaxSize amount.
	maxSize := c.Opt.MaxSize
	if maxSize == 0 {
		// If comparing regular to non-regular file, sizes may not
		// agree...  Use the larger value.
		maxSize = info1.Size()
		if maxSize < info2.Size() {
			maxSize = info2.Size()
		}
		if maxSize == 0 { // possible non-regular files
			maxSize = defaultMaxSize
		}
	}

	if c.multipleMode() {
		h1, err1 := c.getHash(path1, maxSize)
		if err1 != nil {
			return false, err1
		}
		h2, err2 := c.getHash(path2, maxSize)
		if err2 != nil {
			return false, err2
		}
		if !bytes.Equal(h1, h2) {
			return false, nil // hashes mismatch
		}
		// hashes match
		if !c.hashMatchCompare {
			return true, nil // accept hash match without byte-by-byte comparison
		}
		// do byte-by-byte comparison
		c.debugf("CompareFile(%s,%s): hash match, will compare bytes\n", path1, path2)
	}

	// Use our maxSize to avoid triggering the defaultMaxSize for files.
	// We still need to preserve the error returning properties of the
	// input amount exceeding MaxSize, so we can't use LimitedReader.
	c.resetDebugging()

	eq, err := c.compareReader(r1, r2, maxSize)

	c.printDebugCompareReader()

	return eq, err
}

func (c *Cmp) read(r io.Reader, buf []byte) (int, error) {
	n, err := r.Read(buf)

	if err == io.EOF {
		c.debugf("read: EOF found\n")
	}

	if c.Opt.Debug {
		c.readCount++
		c.readSum += int64(n)
		if n < c.readMin {
			c.readMin = n
		}
		if n > c.readMax {
			c.readMax = n
		}
	}

	return n, err
}

// CompareReader verifies that two readers provide same content.
//
// Reading more than MaxSize will return an error (along with the comparison
// value up to MaxSize bytes), unless one or both Readers are LimitedReaders,
// in which case MaxSize is ignored.
func (c *Cmp) CompareReader(r1, r2 io.Reader) (bool, error) {

	c.resetDebugging()

	equal, err := c.compareReader(r1, r2, c.Opt.MaxSize)

	c.printDebugCompareReader()

	return equal, err
}

func (c *Cmp) resetDebugging() {
	if c.Opt.Debug {
		c.readCount = 0
		c.readMin = 2000000000
		c.readMax = 0
		c.readSum = 0
	}
}

func (c *Cmp) printDebugCompareReader() {
	c.debugf("CompareReader(%d,%d): readCount=%d readMin=%d readMax=%d readSum=%d\n",
		len(c.buf), c.Opt.MaxSize, c.readCount, c.readMin, c.readMax, c.readSum)
}

// readPartial keeps reading from reader into provided buffer,
// until buffer size reaches exactly n2. n1 is initial buffer size.
// useful to ensure we get an specific buffer size from reader,
// withstanding partial reads.
func readPartial(c *Cmp, r io.Reader, buf []byte, n1, n2 int) (int, error) {
	for n1 < n2 {
		n, err := c.read(r, buf[n1:n2])
		n1 += n
		if err != nil {
			return n1, err
		}
	}
	return n1, nil
}

func (c *Cmp) compareReader(r1, r2 io.Reader, maxSize int64) (bool, error) {

	// Use LimitedReaders to ensure no data beyond MaxSize or LimitedReader limit
	// (when only one LimitedReader is given) other than at most a single byte.
	var lr1, lr2 io.Reader
	var checkAfterEOF1 bool
	var checkAfterEOF2 bool

	tmpLR1, isLR1 := r1.(*io.LimitedReader)
	tmpLR2, isLR2 := r2.(*io.LimitedReader)

	if isLR1 && isLR2 {
		lr1 = r1
		lr2 = r2
	} else if isLR1 && !isLR2 {
		lr1 = r1
		lr2 = io.LimitReader(r2, tmpLR1.N)
		checkAfterEOF2 = true
	} else if isLR2 && !isLR1 {
		lr2 = r2
		lr1 = io.LimitReader(r1, tmpLR2.N)
		checkAfterEOF1 = true
	} else {
		// Neither Reader is a LimitedReader.  Setup LimitedReaders w/ validated
		// maxSize limit, so that bytes compared will not exceed maxSize.
		if maxSize == 0 {
			maxSize = defaultMaxSize
		}

		if maxSize < 1 {
			return false, fmt.Errorf("nonpositive max size")
		}

		lr1 = io.LimitReader(r1, maxSize)
		lr2 = io.LimitReader(r2, maxSize)
		checkAfterEOF1 = true
		checkAfterEOF2 = true
	}

	buf := c.buf

	size := len(buf) / 2
	if size < 1 {
		return false, fmt.Errorf("insufficient buffer size")
	}

	buf1 := buf[:size]
	buf2 := buf[size : 2*size] // must force same size as buf1

	if len(buf1) != len(buf2) {
		return false, fmt.Errorf("buffer size mismatch buf1=%d buf2=%d", len(buf1), len(buf2))
	}

	eof1 := false
	eof2 := false

	for !eof1 && !eof2 {
		n1, err1 := c.read(lr1, buf1)
		switch err1 {
		case io.EOF:
			eof1 = true
		case nil:
		default:
			return false, err1
		}

		n2, err2 := c.read(lr2, buf2)
		switch err2 {
		case io.EOF:
			eof2 = true
		case nil:
		default:
			return false, err2
		}

		switch {
		case n1 < n2:
			n, errPart := readPartial(c, lr1, buf1, n1, n2)
			switch errPart {
			case io.EOF:
				eof1 = true
			case nil:
			default:
				return false, errPart
			}
			n1 = n
		case n2 < n1:
			n, errPart := readPartial(c, lr2, buf2, n2, n1)
			switch errPart {
			case io.EOF:
				eof2 = true
			case nil:
			default:
				return false, errPart
			}
			n2 = n
		}

		if n1 != n2 {
			c.debugf("compareReader: distinct buffer sizes\n")
			return false, nil
		}

		if !bytes.Equal(buf1[:n1], buf2[:n2]) {
			c.debugf("compareReader: found byte mismatch\n")
			return false, nil
		}
	}

	if !eof1 || !eof2 {
		c.debugf("compareReader: EOF for only one input\n")
		return false, nil
	}

	// Check the EOF status of the original readers. If neither was a
	// LimitedReader, and there is more readable data on either, then
	// return true with an error for exceeding MaxSize.  If one was a
	// LimitedReader and has more data, then return false.  If both were
	// LimitedReaders, then we've reached the desired EOF and return true.
	if checkAfterEOF1 && checkAfterEOF2 {
		// If both original readers need to be checked after EOF, then return
		// 'true' with an error if there is more data in either.
		eof1 = postEOFCheck(c, lr1, buf1[:1])
		eof2 = postEOFCheck(c, lr2, buf2[:1])
		switch {
		case eof1 && eof2:
			return true, nil
		default:
			c.debugf("compareReader: partial match, but max size exceeded\n")
			return true, fmt.Errorf("max read size reached")
		}
	}
	// Return false if only one reader is a LimitedReader, and the other
	// still has data to be read.  Else return true.
	if checkAfterEOF1 {
		return postEOFCheck(c, lr1, buf1[:1]), nil
	}
	if checkAfterEOF2 {
		return postEOFCheck(c, lr2, buf2[:1]), nil
	}

	return true, nil
}

// postEOFCheck returns false if there is more data in a LimitedReader after
// hitting EOF
func postEOFCheck(c *Cmp, r io.Reader, buf []byte) bool {
	tmpLR, isLR := r.(*io.LimitedReader)
	if isLR {
		// If the limit wasn't reached, then we don't need to check for
		// more data after the EOF
		if tmpLR.N > 0 {
			return true
		}

		// Use the internal Reader for checking for more data
		r = tmpLR.R
	} else {
		c.debugf("compareReader: A type assertion of LimitedReader unexpectedly failed\n")
	}

	// Attempt to read more bytes from the original readers, to determine
	// if we should return an error for exceeding the MaxSize read limit.
	n, _ := readPartial(c, r, buf, 0, len(buf))
	return n == 0
}

func (c *Cmp) debugf(format string, v ...interface{}) {
	if c.Opt.Debug {
		fmt.Printf("DEBUG "+format, v...)
	}
}

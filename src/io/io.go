// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package io provides basic interfaces to I/O primitives.
// Its primary job is to wrap existing implementations of such primitives,
// such as those in package os, into shared public interfaces that
// abstract the functionality, plus some other related primitives.
//
// Because these interfaces and primitives wrap lower-level operations with
// various implementations, unless otherwise informed clients should not
// assume they are safe for parallel execution.
package io

import (
	"errors"
	"sync"
)

// Seek whence values.
const (
	SeekStart   = 0 // seek relative to the origin of the file
	SeekCurrent = 1 // seek relative to the current offset
	SeekEnd     = 2 // seek relative to the end
)

// ErrShortWrite means that a write accepted fewer bytes than requested
// but failed to return an explicit error.
var ErrShortWrite = errors.New("short write")

// errInvalidWrite means that a write returned an impossible count.
var errInvalidWrite = errors.New("invalid write result")

// ErrShortBuffer means that a read required a longer buffer than was provided.
// 读取的buffer太小
var ErrShortBuffer = errors.New("short buffer")

// EOF is the error returned by Read when no more input is available.
// (Read must return EOF itself, not an error wrapping EOF,
// because callers will test for EOF using ==.)
// Functions should return EOF only to signal a graceful end of input.
// If the EOF occurs unexpectedly in a structured data stream,
// the appropriate error is either ErrUnexpectedEOF or some other error
// giving more detail.
// Read必须返回未包裹的EOF，避免上层用==判断。固定结构体读取不太应该出现EOR
var EOF = errors.New("EOF")

// ErrUnexpectedEOF means that EOF was encountered in the
// middle of reading a fixed-size block or data structure.
var ErrUnexpectedEOF = errors.New("unexpected EOF")

// ErrNoProgress is returned by some clients of a Reader when
// many calls to Read have failed to return any data or error,
// usually the sign of a broken Reader implementation.
// Read操作不返回数据，Reader实现有问题时出现吗？
var ErrNoProgress = errors.New("multiple Read calls return no data or error")

// Reader is the interface that wraps the basic Read method.
//
// Read reads up to len(p) bytes into p. It returns the number of bytes
// read (0 <= n <= len(p)) and any error encountered. Even if Read
// returns n < len(p), it may use all of p as scratch space during the call.
// If some data is available but not len(p) bytes, Read conventionally
// returns what is available instead of waiting for more.
// 将数据从流读到p，若可读的数据小于p，
//
// When Read encounters an error or end-of-file condition after
// successfully reading n > 0 bytes, it returns the number of
// bytes read. It may return the (non-nil) error from the same call
// or return the error (and n == 0) from a subsequent call.
// An instance of this general case is that a Reader returning
// a non-zero number of bytes at the end of the input stream may
// return either err == EOF or err == nil. The next Read should
// return 0, EOF.
// n表示遇到报错前读到的数据数.
// 一个常见的情况是数据流读完了，返回了一个n<=len(p)和err=EOF，下一次调用必返回0,EOF
//
// Callers should always process the n > 0 bytes returned before
// considering the error err. Doing so correctly handles I/O errors
// that happen after reading some bytes and also both of the
// allowed EOF behaviors.
// 调用者应优先处理读到的n(>0) bytes的数据，再去处理错误
//
// Implementations of Read are discouraged from returning a
// zero byte count with a nil error, except when len(p) == 0.
// Callers should treat a return of 0 and nil as indicating that
// nothing happened; in particular it does not indicate EOF.
//
// Implementations must not retain p.
// 不建议返回(0, nil)这种情况，(0, nil)表示没事发生
// "Implementations must not retain p."指的是p在使用后要销毁掉？
type Reader interface {
	Read(p []byte) (n int, err error)
}

// Writer is the interface that wraps the basic Write method.
//
// Write writes len(p) bytes from p to the underlying data stream.
// It returns the number of bytes written from p (0 <= n <= len(p))
// and any error encountered that caused the write to stop early.
// Write must return a non-nil error if it returns n < len(p).
// Write must not modify the slice data, even temporarily.
//
// Implementations must not retain p.
// 将数据从p写到数据流，当n<len(p)即最后一次调用或遇到报错时时一定要伴随错误码，正常就EOF，不正常就其他
type Writer interface {
	Write(p []byte) (n int, err error)
}

// Closer is the interface that wraps the basic Close method.
//
// The behavior of Close after the first call is undefined.
// Specific implementations may document their own behavior.
// 第一次调用后再调用Close会返回什么是不确定的，要根据实现而定
type Closer interface {
	Close() error
}

// Seeker is the interface that wraps the basic Seek method.
//
// Seek sets the offset for the next Read or Write to offset,
// interpreted according to whence:
// SeekStart means relative to the start of the file,
// SeekCurrent means relative to the current offset, and
// SeekEnd means relative to the end
// (for example, offset = -2 specifies the penultimate byte of the file).
// Seek returns the new offset relative to the start of the
// file or an error, if any.
// Seek用来指定下一次Read和Write方法从什么位置开始
// offset偏移量，whence相对位置，指偏移量相对文件开始、文件当前还是文件结尾
// 返回的是相对"文件开始"的偏移量
//
// Seeking to an offset before the start of the file is an error.
// Seeking to any positive offset may be allowed, but if the new offset exceeds
// the size of the underlying object the behavior of subsequent I/O operations
// is implementation-dependent.
// 早于文件开始是一种错误，任何正的offset都是允许的，超过潜在object大小的行为是依赖实现的（不同实现可能不一样）
type Seeker interface {
	Seek(offset int64, whence int) (int64, error)
}

// ReadWriter is the interface that groups the basic Read and Write methods.
type ReadWriter interface {
	Reader
	Writer
}

// ReadCloser is the interface that groups the basic Read and Close methods.
type ReadCloser interface {
	Reader
	Closer
}

// WriteCloser is the interface that groups the basic Write and Close methods.
type WriteCloser interface {
	Writer
	Closer
}

// ReadWriteCloser is the interface that groups the basic Read, Write and Close methods.
type ReadWriteCloser interface {
	Reader
	Writer
	Closer
}

// ReadSeeker is the interface that groups the basic Read and Seek methods.
type ReadSeeker interface {
	Reader
	Seeker
}

// ReadSeekCloser is the interface that groups the basic Read, Seek and Close
// methods.
type ReadSeekCloser interface {
	Reader
	Seeker
	Closer
}

// WriteSeeker is the interface that groups the basic Write and Seek methods.
type WriteSeeker interface {
	Writer
	Seeker
}

// ReadWriteSeeker is the interface that groups the basic Read, Write and Seek methods.
type ReadWriteSeeker interface {
	Reader
	Writer
	Seeker
}

// ReaderFrom is the interface that wraps the ReadFrom method.
//
// ReadFrom reads data from r until EOF or error.
// The return value n is the number of bytes read.
// Any error except EOF encountered during the read is also returned.
//
// The Copy function uses ReaderFrom if available.
// 从r读数据直到EOF，n为读取到的bytes数，Copy函数用到该接口
type ReaderFrom interface {
	ReadFrom(r Reader) (n int64, err error)
}

// WriterTo is the interface that wraps the WriteTo method.
//
// WriteTo writes data to w until there's no more data to write or
// when an error occurs. The return value n is the number of bytes
// written. Any error encountered during the write is also returned.
//
// The Copy function uses WriterTo if available.
// 写数据到w，n为写入的数据数，Copy函数用到该接口
type WriterTo interface {
	WriteTo(w Writer) (n int64, err error)
}

// ReaderAt is the interface that wraps the basic ReadAt method.
//
// ReadAt reads len(p) bytes into p starting at offset off in the
// underlying input source. It returns the number of bytes
// read (0 <= n <= len(p)) and any error encountered.
// 从off开始读len(p)字节数据到p
//
// When ReadAt returns n < len(p), it returns a non-nil error
// explaining why more bytes were not returned. In this respect,
// ReadAt is stricter than Read.
// 和Read方法的区别是，只要n<len(p)就会返回错误解释为什么。更加严格一些
//
// Even if ReadAt returns n < len(p), it may use all of p as scratch
// space during the call. If some data is available but not len(p) bytes,
// ReadAt blocks until either all the data is available or an error occurs.
// In this respect ReadAt is different from Read.
// 即使ReadAt返回n<len(p), 它仍会用所有p的空间。如果有一些数据是可读的但不是len(p)，有两种表现行为①阻塞等待所有可读的数据读完，或者返回报错
//
// If the n = len(p) bytes returned by ReadAt are at the end of the
// input source, ReadAt may return either err == EOF or err == nil.
// 如果刚好读到结尾，可能返回err == EOF or err == nil
//
// If ReadAt is reading from an input source with a seek offset,
// ReadAt should not affect nor be affected by the underlying
// seek offset.
// ReadAt的行为不应该被潜在的seek offse影响？
//
// Clients of ReadAt can execute parallel ReadAt calls on the
// same input source.
// 在同一个位置可以并行调用ReadAt
//
// Implementations must not retain p.
type ReaderAt interface {
	ReadAt(p []byte, off int64) (n int, err error)
}

// WriterAt is the interface that wraps the basic WriteAt method.
//
// WriteAt writes len(p) bytes from p to the underlying data stream
// at offset off. It returns the number of bytes written from p (0 <= n <= len(p))
// and any error encountered that caused the write to stop early.
// WriteAt must return a non-nil error if it returns n < len(p).
// 和ReadAt类似，n<len(p)必须返回错误
//
// If WriteAt is writing to a destination with a seek offset,
// WriteAt should not affect nor be affected by the underlying
// seek offset.
// 和ReadAt类似，WritteAt的行为不应该被潜在的seek offse影响？
//
// Clients of WriteAt can execute parallel WriteAt calls on the same
// destination if the ranges do not overlap.
// 范围没有重叠是啥意思？
//
// Implementations must not retain p.
type WriterAt interface {
	WriteAt(p []byte, off int64) (n int, err error)
}

// ByteReader is the interface that wraps the ReadByte method.
//
// ReadByte reads and returns the next byte from the input or
// any error encountered. If ReadByte returns an error, no input
// byte was consumed, and the returned byte value is undefined.
// 由于只返回一个byte，报错则表示没有读到数据
//
// ReadByte provides an efficient interface for byte-at-time
// processing. A Reader that does not implement  ByteReader
// can be wrapped using bufio.NewReader to add this method.
// 提供一个逐个byte处理的方法，如果Reader没有实现ByteReader这个接口，可以通过bufio.NewReader来追加
type ByteReader interface {
	ReadByte() (byte, error)
}

// ByteScanner is the interface that adds the UnreadByte method to the
// basic ReadByte method.
//
// UnreadByte causes the next call to ReadByte to return the last byte read.
// If the last operation was not a successful call to ReadByte, UnreadByte may
// return an error, unread the last byte read (or the byte prior to the
// last-unread byte), or (in implementations that support the Seeker interface)
// seek to one byte before the current offset.
// 如果上一个byte读取报错，UnreadByte会返回错误，且①取消最后一个byte的读取并返回到上一次位置？②取消最后一个unread byte之前的byte，和①有啥区别呀？③如果实现了Seeker，也可能会追溯到前一个offset
type ByteScanner interface {
	ByteReader
	UnreadByte() error
}

// ByteWriter is the interface that wraps the WriteByte method.
type ByteWriter interface {
	WriteByte(c byte) error
}

// RuneReader is the interface that wraps the ReadRune method.
//
// ReadRune reads a single encoded Unicode character
// and returns the rune and its size in bytes. If no character is
// available, err will be set.
// 读一个Unicode字符，返回rune=int32字符值（类似c里的ASCII码）和大小和错误码
type RuneReader interface {
	ReadRune() (r rune, size int, err error)
}

// RuneScanner is the interface that adds the UnreadRune method to the
// basic ReadRune method.
//
// UnreadRune causes the next call to ReadRune to return the last rune read.
// If the last operation was not a successful call to ReadRune, UnreadRune may
// return an error, unread the last rune read (or the rune prior to the
// last-unread rune), or (in implementations that support the Seeker interface)
// seek to the start of the rune before the current offset.
// 和ByteScanner类似，只是对象换成了rune
type RuneScanner interface {
	RuneReader
	UnreadRune() error
}

// StringWriter is the interface that wraps the WriteString method.
type StringWriter interface {
	WriteString(s string) (n int, err error)
}

// WriteString writes the contents of the string s to w, which accepts a slice of bytes.
// If w implements StringWriter, its WriteString method is invoked directly.
// Otherwise, w.Write is called exactly once.
// 将数据从s写入w，如果w实现了StringWriter，则替换方法调用
func WriteString(w Writer, s string) (n int, err error) {
	if sw, ok := w.(StringWriter); ok {
		return sw.WriteString(s)
	}
	return w.Write([]byte(s))
}

// ReadAtLeast reads from r into buf until it has read at least min bytes.
// It returns the number of bytes copied and an error if fewer bytes were read.
// The error is EOF only if no bytes were read.
// If an EOF happens after reading fewer than min bytes,
// ReadAtLeast returns ErrUnexpectedEOF.
// EOF表示读完了数据，ErrUnexpectedEOF表示读的数据比min少？感觉Unexpected的说法也不太好
// If min is greater than the length of buf, ReadAtLeast returns ErrShortBuffer.
// 如果min和buf长度不匹配返回ErrShortBuffer
// On return, n >= min if and only if err == nil.
// If r returns an error having read at least min bytes, the error is dropped.
// 当读取的数量大于min就丢弃错误，说明这个函数只保证读取min个byte
func ReadAtLeast(r Reader, buf []byte, min int) (n int, err error) {
	if len(buf) < min {
		return 0, ErrShortBuffer
	}
	for n < min && err == nil {
		var nn int
		nn, err = r.Read(buf[n:])
		n += nn
	}
	// n >= min只可能出现在上一个for循环最后一次循环n+nn>=min
	if n >= min {
		err = nil
	} else if n > 0 && err == EOF {
		err = ErrUnexpectedEOF
	}
	// 当n<min时， n == 0 或者 err != EOF 则用r.Read的错误直接返回
	return
}

// ReadFull reads exactly len(buf) bytes from r into buf.
// It returns the number of bytes copied and an error if fewer bytes were read.
// The error is EOF only if no bytes were read.
// If an EOF happens after reading some but not all the bytes,
// ReadFull returns ErrUnexpectedEOF.
// On return, n == len(buf) if and only if err == nil.
// If r returns an error having read at least len(buf) bytes, the error is dropped.
// 和ReadAtLeast类似，就是调整了下读取下限
func ReadFull(r Reader, buf []byte) (n int, err error) {
	return ReadAtLeast(r, buf, len(buf))
}

// CopyN copies n bytes (or until an error) from src to dst.
// It returns the number of bytes copied and the earliest
// error encountered while copying.
// On return, written == n if and only if err == nil.
//
// If dst implements the ReaderFrom interface,
// the copy is implemented using it.
// 如果dst实现了ReadFrom则直接使用？
func CopyN(dst Writer, src Reader, n int64) (written int64, err error) {
	written, err = Copy(dst, LimitReader(src, n))
	if written == n {
		return n, nil
	}
	if written < n && err == nil {
		// src stopped early; must have been EOF.
		err = EOF
	}
	return
}

// Copy copies from src to dst until either EOF is reached
// on src or an error occurs. It returns the number of bytes
// copied and the first error encountered while copying, if any.
//
// A successful Copy returns err == nil, not err == EOF.
// Because Copy is defined to read from src until EOF, it does
// not treat an EOF from Read as an error to be reported.
//
// If src implements the WriterTo interface,
// the copy is implemented by calling src.WriteTo(dst).
// Otherwise, if dst implements the ReaderFrom interface,
// the copy is implemented by calling dst.ReadFrom(src).
func Copy(dst Writer, src Reader) (written int64, err error) {
	return copyBuffer(dst, src, nil)
}

// CopyBuffer is identical to Copy except that it stages through the
// provided buffer (if one is required) rather than allocating a
// temporary one. If buf is nil, one is allocated; otherwise if it has
// zero length, CopyBuffer panics.
// CopyBuffer和Copy的区别只在于多了个buf参数。buf=nil则重新分配，否则若len(buf)=0则panic
//
// If either src implements WriterTo or dst implements ReaderFrom,
// buf will not be used to perform the copy.
// src和dst实现了WriterTo或ReaderFrom接口的话，buf就没啥用了，会直接写入
func CopyBuffer(dst Writer, src Reader, buf []byte) (written int64, err error) {
	if buf != nil && len(buf) == 0 {
		panic("empty buffer in CopyBuffer")
	}
	return copyBuffer(dst, src, buf)
}

// copyBuffer is the actual implementation of Copy and CopyBuffer.
// if buf is nil, one is allocated.
func copyBuffer(dst Writer, src Reader, buf []byte) (written int64, err error) {
	// If the reader has a WriteTo method, use it to do the copy.
	// Avoids an allocation and a copy.
	if wt, ok := src.(WriterTo); ok {
		return wt.WriteTo(dst)
	}
	// Similarly, if the writer has a ReadFrom method, use it to do the copy.
	if rt, ok := dst.(ReaderFrom); ok {
		return rt.ReadFrom(src)
	}
	if buf == nil {
		// 默认构造32KB的数据，若实现了LimiterReader，则生成l.N大小的buffer
		size := 32 * 1024
		if l, ok := src.(*LimitedReader); ok && int64(size) > l.N {
			if l.N < 1 {
				size = 1
			} else {
				size = int(l.N)
			}
		}
		buf = make([]byte, size)
	}
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			// buf[0:nr]读到的有效数据
			nw, ew := dst.Write(buf[0:nr])
			// 如果写入数据比读到的还多肯定是有问题的，可能写了一些无效数据
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errInvalidWrite
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			// 读到的数据不能完全写入
			if nr != nw {
				err = ErrShortWrite
				break
			}
		}
		// 读报错EOF放过
		if er != nil {
			if er != EOF {
				err = er
			}
			break
		}
	}
	return written, err
}

// LimitReader returns a Reader that reads from r
// but stops with EOF after n bytes.
// The underlying implementation is a *LimitedReader.
func LimitReader(r Reader, n int64) Reader { return &LimitedReader{r, n} }

// A LimitedReader reads from R but limits the amount of
// data returned to just N bytes. Each call to Read
// updates N to reflect the new amount remaining.
// Read returns EOF when N <= 0 or when the underlying R returns EOF.
// 只读取N个数据
// N<=0返回EOF指异常情况吗？
type LimitedReader struct {
	R Reader // underlying reader
	N int64  // max bytes remaining
}

func (l *LimitedReader) Read(p []byte) (n int, err error) {
	if l.N <= 0 {
		return 0, EOF
	}
	if int64(len(p)) > l.N {
		p = p[0:l.N]
	}
	n, err = l.R.Read(p)
	l.N -= int64(n)
	return
}

// NewSectionReader returns a SectionReader that reads from r
// starting at offset off and stops with EOF after n bytes.
func NewSectionReader(r ReaderAt, off int64, n int64) *SectionReader {
	var remaining int64
	const maxint64 = 1<<63 - 1
	if off <= maxint64-n {
		remaining = n + off
	} else {
		// Overflow, with no way to return error.
		// Assume we can read up to an offset of 1<<63 - 1.
		remaining = maxint64
	}
	return &SectionReader{r, off, off, remaining}
}

// SectionReader implements Read, Seek, and ReadAt on a section
// of an underlying ReaderAt.
type SectionReader struct {
	r     ReaderAt
	base  int64
	off   int64
	limit int64
}

func (s *SectionReader) Read(p []byte) (n int, err error) {
	if s.off >= s.limit {
		return 0, EOF
	}
	if max := s.limit - s.off; int64(len(p)) > max {
		p = p[0:max]
	}
	n, err = s.r.ReadAt(p, s.off)
	s.off += int64(n)
	return
}

var errWhence = errors.New("Seek: invalid whence")
var errOffset = errors.New("Seek: invalid offset")

func (s *SectionReader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	default:
		return 0, errWhence
	case SeekStart:
		offset += s.base
	case SeekCurrent:
		offset += s.off
	case SeekEnd:
		offset += s.limit
	}
	if offset < s.base {
		return 0, errOffset
	}
	s.off = offset
	return offset - s.base, nil
}

func (s *SectionReader) ReadAt(p []byte, off int64) (n int, err error) {
	if off < 0 || off >= s.limit-s.base {
		return 0, EOF
	}
	off += s.base
	if max := s.limit - off; int64(len(p)) > max {
		p = p[0:max]
		n, err = s.r.ReadAt(p, off)
		if err == nil {
			err = EOF
		}
		return n, err
	}
	return s.r.ReadAt(p, off)
}

// Size returns the size of the section in bytes.
func (s *SectionReader) Size() int64 { return s.limit - s.base }

// An OffsetWriter maps writes at offset base to offset base+off in the underlying writer.
type OffsetWriter struct {
	w    WriterAt
	base int64 // the original offset
	off  int64 // the current offset
}

// NewOffsetWriter returns an OffsetWriter that writes to w
// starting at offset off.
func NewOffsetWriter(w WriterAt, off int64) *OffsetWriter {
	return &OffsetWriter{w, off, off}
}

func (o *OffsetWriter) Write(p []byte) (n int, err error) {
	n, err = o.w.WriteAt(p, o.off)
	o.off += int64(n)
	return
}

func (o *OffsetWriter) WriteAt(p []byte, off int64) (n int, err error) {
	off += o.base
	return o.w.WriteAt(p, off)
}

func (o *OffsetWriter) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	default:
		return 0, errWhence
	case SeekStart:
		offset += o.base
	case SeekCurrent:
		offset += o.off
	}
	if offset < o.base {
		return 0, errOffset
	}
	o.off = offset
	return offset - o.base, nil
}

// TeeReader returns a Reader that writes to w what it reads from r.
// All reads from r performed through it are matched with
// corresponding writes to w. There is no internal buffering -
// the write must complete before the read completes.
// Any error encountered while writing is reported as a read error.
func TeeReader(r Reader, w Writer) Reader {
	return &teeReader{r, w}
}

type teeReader struct {
	r Reader
	w Writer
}

func (t *teeReader) Read(p []byte) (n int, err error) {
	n, err = t.r.Read(p)
	if n > 0 {
		if n, err := t.w.Write(p[:n]); err != nil {
			return n, err
		}
	}
	return
}

// Discard is a Writer on which all Write calls succeed
// without doing anything.
var Discard Writer = discard{}

type discard struct{}

// discard implements ReaderFrom as an optimization so Copy to
// io.Discard can avoid doing unnecessary work.
var _ ReaderFrom = discard{}

func (discard) Write(p []byte) (int, error) {
	return len(p), nil
}

func (discard) WriteString(s string) (int, error) {
	return len(s), nil
}

var blackHolePool = sync.Pool{
	New: func() any {
		b := make([]byte, 8192)
		return &b
	},
}

func (discard) ReadFrom(r Reader) (n int64, err error) {
	bufp := blackHolePool.Get().(*[]byte)
	readSize := 0
	for {
		readSize, err = r.Read(*bufp)
		n += int64(readSize)
		if err != nil {
			blackHolePool.Put(bufp)
			if err == EOF {
				return n, nil
			}
			return
		}
	}
}

// NopCloser returns a ReadCloser with a no-op Close method wrapping
// the provided Reader r.
// If r implements WriterTo, the returned ReadCloser will implement WriterTo
// by forwarding calls to r.
func NopCloser(r Reader) ReadCloser {
	if _, ok := r.(WriterTo); ok {
		return nopCloserWriterTo{r}
	}
	return nopCloser{r}
}

type nopCloser struct {
	Reader
}

func (nopCloser) Close() error { return nil }

type nopCloserWriterTo struct {
	Reader
}

func (nopCloserWriterTo) Close() error { return nil }

func (c nopCloserWriterTo) WriteTo(w Writer) (n int64, err error) {
	return c.Reader.(WriterTo).WriteTo(w)
}

// ReadAll reads from r until an error or EOF and returns the data it read.
// A successful call returns err == nil, not err == EOF. Because ReadAll is
// defined to read from src until EOF, it does not treat an EOF from Read
// as an error to be reported.
func ReadAll(r Reader) ([]byte, error) {
	b := make([]byte, 0, 512)
	for {
		if len(b) == cap(b) {
			// Add more capacity (let append pick how much).
			b = append(b, 0)[:len(b)]
		}
		n, err := r.Read(b[len(b):cap(b)])
		b = b[:len(b)+n]
		if err != nil {
			if err == EOF {
				err = nil
			}
			return b, err
		}
	}
}

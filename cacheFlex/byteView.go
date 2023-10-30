package cacheFlex

// A ByteView holds an immutable view of bytes.
type ByteView struct {
	bytes []byte
}

// Len returns the view's length
func (v ByteView) Len() int {
	return len(v.bytes)
}

// ByteSlice returns a copy of the data as a byte slice
// bytes inside ByteView struct is read only
// such way we can avoid bytes being modified
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.bytes)
}

// String returns the data as a string, making a copy if necessary.
func (v ByteView) String() string {
	return string(v.bytes)
}

// Helper function called by ByteSlice to make a copy of data
func cloneBytes(bytes []byte) []byte {
	c := make([]byte, len(bytes))
	copy(c, bytes)
	return c
}

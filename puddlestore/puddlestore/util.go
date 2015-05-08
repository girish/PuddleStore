package puddlestore

import (
	"../../tapestry/tapestry"
	"strconv"
	"strings"
)

// Puts the contents of the ID inside the given byte
// Starting at 'start' position
func IdIntoByte(bytes []byte, id *tapestry.ID, start int) {
	for i := 0; i < tapestry.DIGITS; i++ {
		bytes[start+i] = byte(id[i])
	}
}

func ByteIntoAguid(bytes []byte, start uint32) Aguid {
	aguid := ""
	for i := uint32(0); i < tapestry.DIGITS; i++ {
		aguid += strconv.FormatUint(uint64(bytes[start+i]), tapestry.BASE)
	}
	return Aguid(strings.ToUpper(aguid))
}

func MakeZeros(bytes []byte, start uint32) {
	for i := uint32(0); i < tapestry.DIGITS; i++ {
		bytes[start+i] = 0
	}
}

func hashToGuid(id tapestry.ID) Guid {
	s := ""
	for i := 0; i < tapestry.DIGITS; i++ {
		s += strconv.FormatUint(uint64(byte(id[i])), tapestry.BASE)
	}
	return Guid(strings.ToUpper(s))
}

// Helper function used in 'ls'
func makeString(elements [FILES_PER_INODE + 2]string) string {
	ret := ""
	for _, s := range elements {
		if s == "" {
			break
		}
		ret += "\t" + s
	}
	return ret
}

// This is paramount for path hashing
func removeExcessSlashes(path string) string {
	var firstNonSlash, lastNonSlash, start int

	onlySlashes := true
	str := path

	length := len(path)

	// Nothing to do
	if path[0] != '/' && path[length-1] != '/' {
		return str
	}

	// Get the first non slash
	for i := 0; i < length; i++ {
		if str[i] != '/' {
			onlySlashes = false
			firstNonSlash = i
			break
		}
	}

	// Get the last non slash
	for i := length - 1; i >= 0; i-- {
		if str[i] != '/' {
			lastNonSlash = i
			break
		}
	}

	// Guaranteed to be the root path
	if onlySlashes {
		str = "/"
		return str
	} else {
		length = lastNonSlash - firstNonSlash + 1
		if str[0] == '/' {
			start = firstNonSlash - 1
			length++
		} else {
			start = 0
		}

		str = path[start : start+length]
	}

	length = len(str)
	for i := 0; i < length; i++ {
		if i+1 == length {
			break
		}

		if str[i] == '/' && str[i+1] == '/' {
			str = str[:i] + str[i+1:]
			length -= 1
			i -= 1
		}
	}

	return str
}

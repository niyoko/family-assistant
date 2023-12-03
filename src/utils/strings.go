package utils

import "unicode/utf8"

func TrimStringToMaxLen(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	var (
		totalLen     int
		decodedRunes []rune
	)

	for {
		if totalLen >= len(s) {
			break
		}

		r, s := utf8.DecodeRuneInString(s[totalLen:])
		if r == utf8.RuneError {
			break
		}

		if totalLen+s > maxLen {
			break
		}

		totalLen += s
		decodedRunes = append(decodedRunes, r)
	}

	return string(decodedRunes)
}

func ChunkLinesToMaxLen(lines []string, maxLen int) [][]string {
	var (
		chunks   [][]string
		chunk    []string
		chunkLen int
	)

	for _, line := range lines {
		if len(line) > maxLen {
			line = TrimStringToMaxLen(line, maxLen-1)
		}

		if chunkLen+len(line)+1 > maxLen {
			chunks = append(chunks, chunk)
			chunk = nil
			chunkLen = 0
		}

		chunk = append(chunk, line)
		chunkLen += len(line) + 1
	}

	if len(chunk) > 0 {
		chunks = append(chunks, chunk)
	}

	return chunks
}

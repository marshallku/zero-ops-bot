package utils

const MaxMessageLength = 2000

func SplitMessage(content string) []string {
    if len(content) <= MaxMessageLength {
        return []string{content}
    }

    var chunks []string
    runes := []rune(content)

    for len(runes) > 0 {
        end := min(MaxMessageLength, len(runes))

        // Try to split at newline
        if end < len(runes) {
            for i := end - 1; i > end-200 && i > 0; i-- {
                if runes[i] == '\n' {
                    end = i + 1
                    break
                }
            }
        }

        chunks = append(chunks, string(runes[:end]))
        runes = runes[end:]
    }

    return chunks
}

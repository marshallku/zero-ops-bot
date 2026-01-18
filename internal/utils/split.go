package utils

const MaxMessageLength = 1800

func SplitMessage(content string) []string {
    if len(content) <= MaxMessageLength {
        return []string{content}
    }

    var chunks []string
    runes := []rune(content)

    for len(runes) > 0 {
        end := min(MaxMessageLength, len(runes))

        // Split at newline if found
        if end < len(runes) {
            for i := end - 1; i > 0; i-- {
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

package utool

// 示例用法
var jsoncStr = `
{
	// 这是一个单行注释
	"name": "John", /* 这是一个多行注释 */
	"age": 30,
	"description": "This is a string with // a comment inside"
}`

func Jsonc2Json(jsonc []byte) []byte {
	inString := false
	inSingleLineComment := false
	inMultiLineComment := false
	escaped := false
	result := []byte{}

	for i := 0; i < len(jsonc); i++ {
		char := jsonc[i]

		if inString {
			if escaped {
				escaped = false
			} else if char == '\\' {
				escaped = true
			} else if char == '"' {
				inString = false
			}
			result = append(result, char)
		} else if inSingleLineComment {
			if char == '\n' {
				inSingleLineComment = false
				result = append(result, char)
			}
		} else if inMultiLineComment {
			if char == '*' && i+1 < len(jsonc) && jsonc[i+1] == '/' {
				inMultiLineComment = false
				i++
			}
		} else {
			if char == '"' {
				inString = true
				result = append(result, char)
			} else if char == '/' && i+1 < len(jsonc) {
				nextChar := jsonc[i+1]
				if nextChar == '/' {
					inSingleLineComment = true
					i++
				} else if nextChar == '*' {
					inMultiLineComment = true
					i++
				} else {
					result = append(result, char)
				}
			} else {
				result = append(result, char)
			}
		}
	}

	return result
}

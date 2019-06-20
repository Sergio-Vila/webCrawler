package htmlscanner

import "golang.org/x/net/html"

// Compare a byte array and a string without the extra
// copy needed to convert a byte array into a string.
func areEqual(b []byte, s string) bool {
    if len(b) != len(s) {
        return false
    }

    for i, byte := range b {
        if byte != s[i] {
            return false
        }
    }

    return true
}

func tagNameEquals(s string, token *html.Tokenizer) bool {
    tagName, _ := token.TagName()
    return areEqual(tagName, s)
}

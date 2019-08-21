# sqlike
Caches multiple expressions which contain wildcards used with SQL LIKE and finds a pattern which matches a given text

# Supported Wildcards

- [x] Percentage representing 0 or more characters
- [ ] Underscore representing a single character

# Usage

```go
t := sqlike.NewLikeTrie(100)
t.SaveExpression("Hi %, This is message 1", "Meta data 1")
t.SaveExpression("Hi %, This is message 2", "Meta data 2")
t.SaveExpression("This is a third message ending with a %", "Meta data 3")
t.SaveExpression("% this message starts with a wildcard", "Meta data 4")

expr, meta, err := t.FindExpression(
	"This is a third message ending with a this message starts with a wildcard")
if err != nil {
	panic(err)
}
fmt.Printf("Text matched expression: %s\n", expr)
fmt.Printf("Metadata: %v\n", meta)
```
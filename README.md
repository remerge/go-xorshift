# Lock Free XOR Shift for Go

Package `xorshift` provides a pooled, lock-free xorshift+ PRNG source for
`math/rand`.

## Usage

```go
package main

import (
  "fmt"
  rand "github.com/remerge/go-xorshift"
)

func main() {
  fmt.Println(rand.Int63())
}
```

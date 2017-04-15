# go-xorshift

Package `xorshift` provides a pooled, lock-free xorshift+ PRNG source for `math/rand`.

## Install

```bash
go get github.com/remerge/go-xorshift
```

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

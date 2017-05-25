
#### Description
 
* Pharos cache is a lru cache for Go(Golang).
* Thread safety without mutex.

#### Example

The simplest way to use pharos cache:
```go
package main

import (
	"github.com/pharosnet/pharoscache"
	"fmt"
)

func main() {
        c := pharoscache.NewCache(&pharoscache.Option{MaxEntries:10})
        c.Bucket("name").Set("key", []byte("value"), 0)
        fmt.Println(c.Bucket("name").Get("key"))
        c.Close()	
}
```


#### Thread safety

Pharos cache is protected by chan for concurrent getters and setters.


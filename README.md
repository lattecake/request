# 使用方式

```
$ go get github.com/lattecake/request
```

### application/x-www-form-urlencoded

```go
import (
	"fmt"
	"github.com/lattecake/request"
	"net/url"
)

func main(){
	r := request.NewRequest()
	params := url.Values{}
	params.Set("a", "b")
	res, err, _ := r.Post("https://lattecake.com", params, nil, "")
	if err != nil {
		// error
		panic(err)
	}
	
	fmt.Println(string(res))
}
```

### application/json

```go
import (
	"fmt"
	"github.com/lattecake/request"
	"net/url"
)

func main(){
	r := request.NewRequest()
	params := url.Values{}
	params.Set("", `{"a":"b"}`)
	res, err, _ := r.Post("https://lattecake.com", params, map[string]string{
		"Context-Type": "application/json",
	}, "")
	if err != nil {
		// error
		panic(err)
	}

	fmt.Println(string(res))
}
```


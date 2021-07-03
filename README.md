# spa
[Gin](https://github.com/gin-gonic/gin) middleware for serving spa development server and spa static files.

## Installation

    go get github.com/frozturk/spa

## Example
```go
package main

import (
	"github.com/frozturk/spa"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	//your api
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	if gin.Mode() == gin.ReleaseMode {
		//serve spa static files
		r.Use(spa.UseSpaStaticFiles(spa.Config{
			SPADirectory: "client/dist",
		}))
	} else {
		//serve angular dev server
		r.Use(spa.UseAngularCliServer(spa.Config{
			SPADirectory: "client",
 			//NPMScript: "start",
		}))
	}

	r.Run()
}
```
## TODO
- [ ] Add react dev server support
## Credits
[WebsocketProxy](https://github.com/koding/websocketproxy)
